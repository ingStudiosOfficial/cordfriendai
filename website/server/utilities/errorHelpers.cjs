function sendErrorResponse(res, statusCode, message, errors = null) {
    return res.status(statusCode).json({
        status: 'error',
        message: message,
        errors: errors 
    });
}

module.exports = { sendErrorResponse };