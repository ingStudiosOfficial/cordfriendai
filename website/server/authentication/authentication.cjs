const path = require('path');
const bcrypt = require('bcrypt');
const jwt = require('jsonwebtoken');
const { ObjectId } = require('mongodb');
require('dotenv').config({ path: path.join(__dirname, '../.env') });

// bcrypt setup
const bcryptSaltRounds = 10;

// JWT setup
const jwtSecretKey = process.env.JWT_SECRET_KEY;

function hashPassword(passwordToHash) {
	return new Promise((resolve, reject) => {
        bcrypt.hash(passwordToHash, bcryptSaltRounds, (err, hash) => {
            if (err) {
                console.log('Failed to hash password:', err);
                return reject(err);
            }
            resolve(hash);
        });
    });
}

function authenticateToken(collection) {
    return async (req, res, next) => {
        const token = req.cookies.auth_token;

        if (!token) {
            return res.status(401).json({ message: 'Access token required.' });
        }

        jwt.verify(token, jwtSecretKey, async (err, decoded) => {
            if (err) {
                return res.status(401).json({ message: 'Invalid or expired token.' });
            }

            console.log('User authenticated:', decoded);

            try {
                console.log('Decoded user ID:', decoded.id);

                const userFromDatabase = await collection.findOne({ 
                    '_id': new ObjectId(decoded.id)
                });

                console.log('User from database:', userFromDatabase);

                if (!userFromDatabase || userFromDatabase.tokenVersion !== decoded.tokenVersion) {
                    return res.status(401).json({ message: 'Token is no longer valid.' });
                }

                req.user = decoded;
                next();
            } catch (error) {
                console.error('Error fetching user:', error);
                return res.status(500).json({ message: 'Internal server error.' });
            }
        });
    };
}

module.exports = { hashPassword, authenticateToken }