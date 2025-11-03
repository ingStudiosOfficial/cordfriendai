function checkFileType(req, file, cb) {
    if (file.mimetype === 'image/png' || 
        file.mimetype === 'image/jpeg' || 
        file.mimetype === 'image/jpg' || 
        file.mimetype === 'image/webp') 
    {
        // Accept the file
        cb(null, true);
    } else {
        // Reject the file
        cb(new Error('Only PNG, JPG, JPEG, and WebP image formats are allowed.'), false);
    }
}

module.exports = { checkFileType };