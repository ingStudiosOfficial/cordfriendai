const express = require('express');
const cors = require('cors');
const fetch = (...args) => import('node-fetch').then(({default: fetch}) => fetch(...args));
const path = require('path');
require('dotenv').config({ path: path.join(__dirname, '../.env') });
const { MongoClient, ObjectId, GridFSBucket } = require('mongodb');
const { Readable } = require('stream');
const multer = require('multer');
const jwt = require('jsonwebtoken');
const cookieParser = require('cookie-parser');

// Import utilities
const { hashPassword, authenticateToken } = require('./utilities/authentication.cjs');
const { checkFileType } = require('./utilities/filetype.cjs');
const { sendErrorResponse } = require('./utilities/errorHelpers.cjs');
const { encryptApiKey, decryptApiKey } = require('./utilities/apiEncryption.cjs');
const { deleteBot } = require('./utilities/botUtilities.cjs');

// Import middleware
const { validateUser } = require('./middleware/validateUser.cjs');
const { validateBotInput } = require('./middleware/validateBot.cjs');
const { validateUserInput } = require('./middleware/validateUserInput.cjs');

const app = express();

app.use(cors({
	origin: process.env.CLIENT_URL,
	credentials: true,
	methods: ['GET', 'POST', 'PUT', 'DELETE']
}));
app.use(cookieParser());
app.use(express.json());

// MongoDB setup
const mongodbUri = process.env.MONGODB_CONNECTION_STRING;
const mongodbDatabaseName = 'cordfriendAI';
const mongoClientOptions = {
    serverSelectionTimeoutMS: 60000, // Timeout after 60s of trying to find a server
    socketTimeoutMS: 60000,          // Timeout after 60s of inactivity on a socket
};
console.log('MongoDB connection string:', mongodbUri);

// JWT setup
const jwtSecretKey = process.env.JWT_SECRET_KEY;

// Port setup
const port = process.env.PORT;
console.log('Port to listen:', port);

// Multer setup - stores the file temporarily in memory
const storage = multer.memoryStorage();
const upload = multer({
	storage: storage,
	limits: {
		fileSize: 5 * 1024 * 1024 // 5 MB file size limit
	},
	fileFilter: checkFileType
});

let database;

// The DB collections
let usersCollection;
let botsCollection;

// The buckets
let botImagesBucket;

async function pingMongodb() {
	const client = new MongoClient(mongodbUri);

	try {
		await client.connect();
		await client.db("admin").command({ ping: 1 });
        console.log("Connected successfully to server and ping successful.");
	} catch (error) {
		console.error('Error pinging MongoDB:', error);
	} finally {
		client.close();
	}
}

async function connectToMongodb() {
	console.log('Attempting to connect to MongoDB...');
	const client = new MongoClient(mongodbUri, mongoClientOptions);

	try {
		await client.connect();
		console.log('Successfully connected to MongoDB!');

		database = client.db(mongodbDatabaseName);

		// Initialize all the collections
		usersCollection = database.collection('users');
		botsCollection = database.collection('bots');

		// Initialize all the buckets
		botImagesBucket = new GridFSBucket(database, {
			bucketName: 'bot_images'
		});
	} catch (error) {
		console.error('MongoDB connection failed:', error);
		process.exit(1);
	}
}

connectToMongodb().then(() => {
	app.listen(port, () => { console.log(`App is listening on port ${port}.`) });

	app.get('/api/verify-auth/', authenticateToken(usersCollection), (req, res) => {
		console.log('User authenticated with valid JWT token, responding.');
		res.status(204).end();
	});

	app.post('/api/login/', async (req, res) => {
		if (!usersCollection) {
			return sendErrorResponse(res, 503, 'The users collection is unavailable.');
		}

		const userData = req.body;

		if (!userData.email || !userData.password) {
			return sendErrorResponse(res, 400, 'Please provide an email or a password.');
		}

		const userEmail = userData.email;
		const userPassword = userData.password;
		
		console.log('User to fetch:', userEmail);

		try {
			console.log(`Attempting to fetch ${userEmail}...`)
			const documentToFetch = await usersCollection.findOne({ email: userEmail });
			console.log('Fetch complete.');

			if (!documentToFetch) {
				return sendErrorResponse(res, 404, `Your email ${userEmail} cannot be found in our database.`);
			}

			if (!documentToFetch.email || !documentToFetch.password) {
				console.log('User cannot be found in the database.');
				return sendErrorResponse(res, 401, 'Your email or password could not be found in our database.');
			}

			// Check and insert token version
			if (!documentToFetch.tokenVersion) {
				const updateResult = await usersCollection.updateOne(
					{ '_id': documentToFetch._id },
					{ $set: { 'tokenVersion': 0 } }
				);

				if (updateResult.matchedCount === 0) {
					console.error('Could not find user when updating token version.');
					return sendErrorResponse(res, 404, 'Could not find your account while updating token.');
				}
			}

			console.log('User found, authenticating...');
			const userValidation = await validateUser(userPassword, documentToFetch.password);
			if (userValidation === false) {
				console.log('Password is incorrect.');
				return sendErrorResponse(res, 401, 'The password you entered is incorrect.');
			}

			const payload = {
				id: documentToFetch._id,
				email: documentToFetch.email,
				tokenVersion: documentToFetch.tokenVersion
			};

			console.log('Payload:', payload);

			const DURATION_DAYS = 7;
			const cookieAgeMs = DURATION_DAYS * 24 * 60 * 60 * 1000; 
			const tokenExpiry = `${DURATION_DAYS}d`;

			const token = jwt.sign(payload, jwtSecretKey, { expiresIn: tokenExpiry });

			res.status(200).cookie('auth_token', token, {
				httpOnly: true,
				secure: process.env.NODE_ENV === 'production',
				sameSite: process.env.NODE_ENV === 'production' ? 'none' : 'lax',
				maxAge: cookieAgeMs
			}).json({
				'message': 'You have successfully logged in, redirecting...'
			});
		} catch (error) {
			console.error('Error while fetching user:', error);
			sendErrorResponse(res, 500, 'An internal server error occurred.', error);
		}
	});

	app.post('/api/signup/', async (req, res) => {
		if (!usersCollection) {
			return sendErrorResponse(res, 503, 'The users collection is unavailable.');
		}

		const userData = req.body;
		console.log('User data:', userData);

		if (!userData.email || !userData.password) {
			return sendErrorResponse(res, 400, 'Please provide an email or a password.');
		}

		const userEmail = userData.email;
		console.log('Creating user:', userEmail);

		if (await usersCollection.findOne({ email: userEmail })) {
			return sendErrorResponse(res, 409, 'User already exists.');
		}

		let userPassword;
		try {
			userPassword = await hashPassword(userData.password);
			console.log('Hashed user password created.');
		} catch (error) {
			console.error('Error while hashing password:', error);
			return sendErrorResponse(res, 500, 'An internal server error occurred while proccessing the passwrod.', error);
		}

		const userDataToStore = {
			'email': userEmail,
			'password': userPassword,
			'bots': [],
			'tokenVersion': 0
		};

		try {
			const result = await usersCollection.insertOne(userDataToStore);

			res.status(201).json({
				'message': 'You have successfully signed up!'
			});
		} catch (error) {
			console.error('Error while creating user:', error);
			sendErrorResponse(res, 500, 'An internal server error occurred.', error);
		}
	});

	app.post('/api/logout/', async (req, res) => {
		const token = req.cookies.auth_token;
		
		if (token) {
			try {
				// Try to verify and invalidate
				const decoded = jwt.verify(token, process.env.JWT_SECRET_KEY);
				
				await usersCollection.updateOne(
					{ '_id': new ObjectId(decoded.id) },
					{ $inc: { tokenVersion: 1 } }
				);
				
				console.log(`User ${decoded.id} logged out.`);
			} catch (error) {
				// Token invalid - clear cookie too
				console.log('Logout with invalid token:', error.message);
			}
		}
		
		// Always clear the cookie
		res.clearCookie('auth_token', {
			httpOnly: true,
			secure: process.env.NODE_ENV === 'production',
			sameSite: process.env.NODE_ENV === 'production' ? 'none' : 'lax',
			path: '/'
		});
		
		res.status(200).json({ message: 'Logged out successfully' });
	});

	app.get('/api/user/get/', authenticateToken(usersCollection), async (req, res) => {
		if (!usersCollection) {
			return sendErrorResponse(res, 503, 'The users collection is unavailable.');
		}

		if (!ObjectId.isValid(req.user.id)) {
			console.log('Failed to fetch user data - invalid format:', req.user.id);
			return sendErrorResponse(res, 400, 'Invalid user ID format.');
		}

		try {
			const userId = new ObjectId(req.user.id);

			const fetchedUser = await usersCollection.findOne({ '_id': userId });

			const payload = {
				id: fetchedUser._id,
				email: fetchedUser.email
			};

			console.log('Returning payload:', payload);

			res.status(200).json({
				'message': 'User fetch successful.',
				'user': payload
			});
		} catch (error) {
			console.error('An error occurred while fetching user data:', error);
			sendErrorResponse(res, 500, 'An internal server error occurred.', error);
		}
	});

	app.put('/api/user/edit/', authenticateToken(usersCollection), validateUserInput, async (req, res) => {
		if (!usersCollection) {
			return sendErrorResponse(res, 503, 'The users collection is unavailable.');
		}

		if (!ObjectId.isValid(req.user.id)) {
			return sendErrorResponse(res, 400, 'Invalid user ID format.');
		}

		try {
			const userId = new ObjectId(req.user.id);

			console.log('Request body:', req.body);
			const userData = req.body;
			console.log('Saving user data:', userData);

			const originalUserData = await usersCollection.findOne({ '_id': userId });
			console.log('Original user data:', originalUserData);

			if (!originalUserData) {
				return sendErrorResponse(res, 404, 'User not found.');
			}

			// Check if the changed email is not the email from the database
			if (userData.email !== originalUserData.email) {
				// Validate if new email exists already
				const emailToFind = await usersCollection.findOne({ 'email': userData.email });

				if (emailToFind) {
					return sendErrorResponse(res, 409, 'User already exists.');
				}
			}

			// Password change logic
			if (userData.new_password) {
				// Validate password
				const passwordValidation = await validateUser(userData.old_password, originalUserData.password);

				if (passwordValidation === false) {
					return sendErrorResponse(res, 401, 'The password you entered is incorrect.');
				}

				// Only if new_password exists and validation passed
				if (userData.new_password) { 
					userData.password = await hashPassword(userData.new_password); 
					delete userData.new_password;
				}

				delete userData.old_password;
			}

			// Iterate over each value to update
			Object.keys(userData).forEach((userDataKey) => {
				originalUserData[userDataKey] = userData[userDataKey];
			});

			const result = await usersCollection.updateOne(
				{ '_id': userId },
				{ $set: userData }
			);

			if (result.matchedCount === 0) {
				return sendErrorResponse(res, 404, 'User not found.');
			}

			res.status(200).json({
				'message': 'User updated successfully.'
			});
		} catch (error) {
			console.error('An error occurred while updating user:', error);
			sendErrorResponse(res, 500, 'An internal server error occurred.', error);
		}
	});

	app.delete('/api/user/delete/', authenticateToken(usersCollection), async (req, res) => {
		if (!usersCollection) {
			return sendErrorResponse(res, 503, 'The users collection is unavailable.');
		}

		if (!botsCollection) {
			return sendErrorResponse(res, 503, 'The bots collection is unavailable.');
		}

		if (!ObjectId.isValid(req.user.id)) {
			return sendErrorResponse(res, 400, 'Invalid user ID format.');
		}

		try {
			const userId = new ObjectId(req.user.id);

			const userResult = await usersCollection.findOne({ '_id': userId });

			if (!userResult) {
				return sendErrorResponse(res, 404, `Could not find the user ${req.user.id}.`);
			}

			if (userResult.bots.length !== 0) {
				console.log(`Deleting ${userResult.bots.length} bots...`);
				
				for (let i = 0; i < userResult.bots.length; i++) {
					const bot = userResult.bots[i];
					console.log(`Attempting to delete bot ${i + 1} out of ${userResult.bots.length}...`);

					if (!ObjectId.isValid(bot)) {
						console.warn(`Bot ${i + 1} has an invalid object ID:`, bot);
						continue;
					}

					try {
						const botToDelete = await botsCollection.findOne({ '_id': new ObjectId(bot) });

						if (!botToDelete) {
							console.warn(`Failed to find bot ${i + 1}:`, bot);
							continue;
						}

						const botId = botToDelete._id;
						const imageId = botToDelete.image_id;

						if (!botId || !imageId) {
							console.warn('Bot or image ID not available for bot:', botToDelete);
							continue;
						}

						await deleteBot(botsCollection, usersCollection, botImagesBucket, botId, userId, imageId, false);
						console.log(`Successfully deleted bot ${i + 1}`);
					} catch (error) {
						console.error(`Error deleting bot ${i + 1}:`, error);
					}
				}
			}

			// Delete the user
			const deletedResult = await usersCollection.deleteOne({ '_id': userId });

			if (deletedResult.deletedCount === 0) {
				return sendErrorResponse(res, 404, `Could not find the user ${req.user.id} while deleting.`);
			}

			// Clear the auth token
			res.clearCookie('auth_token', {
				httpOnly: true,
				secure: process.env.NODE_ENV === 'production',
				sameSite: process.env.NODE_ENV === 'production' ? 'none' : 'lax',
				path: '/'
			});

			res.status(200).json({
				'message': 'User deleted successfully.'
			});
		} catch (error) {
			console.error('Error deleting user:', error);
			sendErrorResponse(res, 500, 'An internal server error occurred.', error);
		}
	});

	app.post('/api/bot/create/', authenticateToken(usersCollection), validateBotInput, async (req, res) => {
		if (!botsCollection) {
			return sendErrorResponse(res, 503, 'The bots collection is unavailable.');
		}

		if (!usersCollection) {
			return sendErrorResponse(res, 503, 'The users collection is unavailable.');
		}

		const botData = req.body;
		console.log('Bot data:', botData);

		if (!botData) {
			return sendErrorResponse(res, 400, 'No bot data provided.');
		}

		if (await botsCollection.findOne({ 'server_id': botData.server_id })) {
			return sendErrorResponse(res, 409, 'Bot already exists in server.');
		}

		const botDataToStore = {
			'name': botData.name,
			'server_id': botData.server_id,
			'user_id': botData.user_id,
			'google_ai_api': encryptApiKey(botData.google_ai_api),
			'image_id': botData.image_id,
			'image_filename': botData.image_filename
		};

		console.log('Bot data to store:', botDataToStore);

		let insertedBotId;

		try {
			const result = await botsCollection.insertOne(botDataToStore);
			insertedBotId = result.insertedId;
			console.log('Bot created:', insertedBotId);
		} catch (error) {
			console.error('Error creating bot:', error);
			return sendErrorResponse(res, 500, 'An internal server error occurred, please try again.', error);
		}

		const userId = new ObjectId(req.user.id);
		console.log('User ID to add:', userId);

		try {
			const result = await usersCollection.updateOne(
				{ _id: userId },
				{ $push: { 'bots': insertedBotId } }
			);

			console.log('Successfully appended bot to user bots');

			res.status(201).json({
				'message': 'Bot successfully created and added to your account.',
				'botId': insertedBotId
			});
		} catch (error) {
			console.error('An error occurred while adding bot to user:', error);
			sendErrorResponse(res, 500, 'An internal server error occurred, please try again.', error);
		}
	});

	app.get('/api/bot/get/all/', authenticateToken(usersCollection), async (req, res) => {
		if (!botsCollection) {
			return sendErrorResponse(res, 503, 'The bots collection is unavailable.');
		}

		if (!usersCollection) {
			return sendErrorResponse(res, 503, 'The users collection is unavailable.');
		}

		let bots;
		try {
			const fetchedUser = await usersCollection.findOne({ '_id': new ObjectId(req.user.id) });

			if (!fetchedUser) {
				console.error(`User with ID ${req.user.id} not found.`);
				return sendErrorResponse(res, 404, 'User not found.');
			}

			bots = fetchedUser.bots;
			console.log('Bots from user fetch:', bots);
		} catch (error) {
			console.error('Error while fetching user:', error);
			return sendErrorResponse(res, 500, 'An internal server error occurred.', error);
		}

		try {
			// Use Promise.all to loop through each bot
			const fetchedBots = await Promise.all(
				bots.map(async (bot, index) => {
					// Convert the string to an ObjectId first
					bot = new ObjectId(bot);

					console.log(`Fetching bot ${index + 1} of ${bots.length}:`, bot);
					const fetchedBot = await botsCollection.findOne({ '_id': bot });
					
					if (!fetchedBot) {
						console.warn(`Bot with ID ${bot} not found.`);
						return null;
					}

					// Decrypt the API key
					fetchedBot.google_ai_api = decryptApiKey(fetchedBot.google_ai_api);

					console.log('Fetched bot:', fetchedBot);
					return fetchedBot;
				})
			);

			// Filter out null values (missing bots)
			const validBots = fetchedBots.filter(bot => bot !== null);

			if (validBots.length !== bots.length) {
				res.status(200).send({
					'message': 'Some bots were not fetched.',
					'bots': validBots
				});
			} else {
				res.status(200).send({
					'message': 'All bots were successfully fetched.',
					'bots': validBots
				});
			}
		} catch (error) {
			console.error('An error occurred while fetching all bots:', error);
			return sendErrorResponse(res, 500, 'Failed to fetch bots.');
		}
	});

	app.get('/api/bot/get/:id', authenticateToken(usersCollection), async (req, res) => {
		if (!botsCollection) {
			return sendErrorResponse(res, 503, 'The bots collection is unavailable.');
		}

		const botId = req.params.id;

		if (!ObjectId.isValid(botId)) {
				return sendErrorResponse(res, 400, 'Invalid bot ID format.');
		}

		try {
			const fetchedBot = await botsCollection.findOne({ id: botId });

			// Decrypt the API key
			fetchedBot.google_ai_api = decryptApiKey(fetchedBot.google_ai_api);

			console.log('Fetched bot:', fetchedBot);

			if (!fetchedBot) {
				return sendErrorResponse(res, 404, `Bot with ID ${botId} not found.`);
			}

			res.status(200).json({
				'bot': fetchedBot
			});
		} catch (error) {
			console.error('An error occurred while fetching bot:', error);
			sendErrorResponse(res, 500, 'An internal server error occurred.', error);
		}
	});

	app.put('/api/bot/edit/:id', authenticateToken(usersCollection), validateBotInput, async (req, res) => {
		if (!botsCollection) {
			return sendErrorResponse(res, 503, 'The bots collection is unavailable.');
		}

		const botData = req.body;
		console.log('Bot data:', botData);

		// Encrypt API key
		botData.google_ai_api = encryptApiKey(botData.google_ai_api);

		if (!botData) {
			return sendErrorResponse(res, 400, 'No bot data provided.');
		}

		// Validate the ID format
		if (!ObjectId.isValid(req.params.id)) {
			return sendErrorResponse(res, 400, 'Invalid bot ID format.');
		}

		// Validate the image and old image ID formats
		if (!ObjectId.isValid(req.body.image_id)) {
			return sendErrorResponse(res, 400, 'Invalid images ID format.');
		}

		if (!ObjectId.isValid(req.body.old_image_id)) {
			return sendErrorResponse(res, 400, 'Invalid old image ID format.');
		}

		const botId = botData._id;

		if (botId && ObjectId.isValid(botId)) {
			const objectId = new ObjectId(botId);
			
			const duplicateBot = await botsCollection.findOne({
				'server_id': botData.server_id,
				'_id': { $ne: objectId } 
			});

			if (duplicateBot) {
				return sendErrorResponse(res, 409, 'Bot already exists in server.');
			}
		}

		try {
			// Delete the _id from being set
			delete botData._id;

			const result = await botsCollection.updateOne(
				{ _id: new ObjectId(req.params.id) },
				{ $set: botData }
			);

			// Check that the bot was found
			if (result.matchedCount === 0) {
				return sendErrorResponse(res, 404, 'Bot not found.');
			}
			
			// Delete previous image from GridFS
			if (req.body.image_id !== req.body.old_image_id) {
				await botImagesBucket.delete(new ObjectId(req.body.old_image_id));
			}

			console.log(`Updated bot with ID ${req.params.id} successfully.`);
			res.status(200).json({
				message: 'Bot successfully updated.'
			});
		} catch (error) {
			console.error('Error while updating bot:', error);
			sendErrorResponse(res, 500, 'An internal server error occurred.', error);
		}
	});

	app.delete('/api/bot/delete/', authenticateToken(usersCollection), async (req, res) => {
		if (!botsCollection) {
			return sendErrorResponse(res, 503, 'The bots collection is unavailable.');
		}

		if (!usersCollection) {
			return sendErrorResponse(res, 503, 'The users collection is unavailable.');
		}

		const botData = req.body;
		console.log('Bot data:', botData);

		try {
			const result = await deleteBot(botsCollection, usersCollection, botImagesBucket, botData._id, req.user.id, botData.image_id);
			res.status(200).json(result);
		} catch (error) {
			console.log('An error occurred while deleting bot:', error);
			
			// Map error messages to appropriate status codes
			if (error.message.includes('Invalid') || error.message.includes('format')) {
				return sendErrorResponse(res, 400, error.message);
			}
			if (error.message.includes('not found')) {
				return sendErrorResponse(res, 404, error.message);
			}

			return sendErrorResponse(res, 500, 'An internal server error occurred.', error);
		}
	});

	app.post('/api/bot/image-upload/', authenticateToken(usersCollection), upload.single('bot-profile-picture'), (req, res) => {
		if (!req.file) {
			return sendErrorResponse(res, 400, 'Please upload an image.');
		}

		const readablePhotoStream = new Readable();
		const fileBuffer = req.file.buffer; 
		
		// Complete the Readable Stream
		readablePhotoStream.push(fileBuffer); 
		readablePhotoStream.push(null); 

		// Open the GridFS destination stream
		const uploadStream = botImagesBucket.openUploadStream(req.file.originalname, {
			contentType: req.file.mimetype
		});

		// Pipe the data - connects the Readable Stream to a Writeable Stream, which is used to write data into the bot collection
		readablePhotoStream.pipe(uploadStream);

		// Handle error events
		uploadStream.on('error', (error) => {
			console.error('GridFS Upload Error:', error);
			return sendErrorResponse(res, 500, 'Failed to upload image.', error);
		});

		uploadStream.on('finish', () => {
			const fileId = uploadStream.id; 
			const filename = uploadStream.filename;
			
			// Return the ID of the image on success
			res.status(201).json({
				message: 'Image successfully uploaded and staged.',
				fileId: fileId,
				filename: filename
			});
		});
	});

	app.get('/api/bot/image-download/:id', async (req, res) => {
		console.log('Image download requested for ID:', req.params.id);

		if (!ObjectId.isValid(req.params.id)) {
			return sendErrorResponse(res, 400, 'Invalid image ID format.');
		}
		
		try {
			const bucket = new GridFSBucket(database, { bucketName: 'bot_images' });
			
			// Verify the file exists first
			const files = await bucket.find({ _id: new ObjectId(req.params.id) }).toArray();
			console.log('Found files:', files);
			
			if (files.length === 0) {
				console.log('No file found with that ID');
				return sendErrorResponse(res, 404, 'Image not found.');
			}
			
			const downloadStream = bucket.openDownloadStream(new ObjectId(req.params.id));

			if (files.length > 0 && files[0].contentType) {
				res.set('Content-Type', files[0].contentType); // Use stored content type
			} else {
				res.set('Content-Type', 'image/*'); // Generic image type
			}

			res.set('Cache-Control', 'public, max-age=86400');

			downloadStream.on('error', (error) => {
				console.error('Download stream error:', error);
				if (!res.headersSent) {
					return sendErrorResponse(res, 404, 'Image to download not found.');
				}
			});
			
			downloadStream.pipe(res);
		} catch (error) {
			console.error('Error in image download:', error);
			if (!res.headersSent) {
				sendErrorResponse(res, 500, 'An error occurred while downloading the image.');
			}
		}
	});
});