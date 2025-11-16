const GoogleStrategy = require('passport-google-oidc').Strategy;
const jwt = require('jsonwebtoken');
const path = require('path');
require('dotenv').config({ path: path.join(__dirname, '../../.env') });

function createGoogleStrategy(usersCollection, credsCollection) {
    return new GoogleStrategy({
        clientID: process.env.GOOGLE_CLIENT_ID,
        clientSecret: process.env.GOOGLE_CLIENT_SECRET,
        callbackURL: '/api/oauth2/callback/google/',
        scope: ['profile', 'email']
    }, async function verify(issuer, profile, cb) {
        try {
            const credential = await credsCollection.findOne({
                provider: issuer,
                subject: profile.id
            });

            if (!credential) {
                if (!profile.emails?.[0]?.value) {
                    console.error('Error while fetching email.');
                    return cb(new Error('Error while fetching email from Google.'));
                }

                const userEmail = profile.emails[0].value;

                const userDataToStore = {
                    'email': userEmail,
                    'password': null,
                    'bots': [],
                    'tokenVersion': 0
                };

                const result = await usersCollection.insertOne(userDataToStore);

                if (!result.insertedId) {
                    return cb(new Error('Error while inserting user.'));
                }

                const userId = result.insertedId;

                const credsResult = await credsCollection.insertOne({
                    userId: userId,
                    provider: issuer,
                    subject: profile.id
                });

                if (!credsResult) {
                    return cb(new Error('Error while inserting credentials.'));
                }

                const payload = {
                    id: userId,
                    email: userDataToStore.email,
                    tokenVersion: userDataToStore.tokenVersion
                };

                const jwtSecretKey = process.env.JWT_SECRET_KEY;
                
                const DURATION_DAYS = 7;
                const cookieAgeMs = DURATION_DAYS * 24 * 60 * 60 * 1000; 
                const tokenExpiry = `${DURATION_DAYS}d`;

                const token = jwt.sign(payload, jwtSecretKey, { expiresIn: tokenExpiry });

                return cb(null, { token });
            }

            const fetchedUser = await usersCollection.findOne({ _id: credential.userId });

            if (!fetchedUser) {
                return cb(new Error('User does not exist.'));
            }

            const payload = {
                id: fetchedUser._id,
                email: fetchedUser.email,
                tokenVersion: fetchedUser.tokenVersion
            };

            const jwtSecretKey = process.env.JWT_SECRET_KEY;
            
            const DURATION_DAYS = 7;
			const tokenExpiry = `${DURATION_DAYS}d`;

            const token = jwt.sign(payload, jwtSecretKey, { expiresIn: tokenExpiry });

            return cb(null, { token });
        } catch (error) {
            console.error('Error while fetching Google OAuth:', error);
            return cb(error);
        }
    });
}

module.exports = { createGoogleStrategy };