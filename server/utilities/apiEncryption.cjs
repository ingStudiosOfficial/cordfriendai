const path = require('path');
const crypto = require('crypto');
require('dotenv').config({ path: path.join(__dirname, '../.env') });

const algorithm = 'aes-256-cbc';
const secretKey = Buffer.from(process.env.CRYPTO_SECRET_KEY, 'hex'); // Convert to Buffer
const iv = Buffer.from(process.env.CRYPTO_IV, 'hex'); // Convert to Buffer

function encryptApiKey(key) {
    const cipher = crypto.createCipheriv(algorithm, secretKey, iv);
    let encrypted = cipher.update(key, 'utf-8', 'hex');
    encrypted += cipher.final('hex');
    return { iv: iv.toString('hex'), encryptedData: encrypted };
}

function decryptApiKey(encryptedKey) {
    const decipher = crypto.createDecipheriv(algorithm, secretKey, Buffer.from(encryptedKey.iv, 'hex'));
    let decrypted = decipher.update(encryptedKey.encryptedData, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    return decrypted;
}

module.exports = { encryptApiKey, decryptApiKey };