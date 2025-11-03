const botUpdateSchema = require('../schemas/botSchema.cjs');

function validateBotInput(req, res, next) {
    const validationOptions = {
        abortEarly: false,
        allowUnknown: false,
        stripUnknown: true
    };

    // Run the validation schema against the request body
    const { error, value } = botUpdateSchema.validate(req.body, validationOptions);

    if (error) {
        const validationErrors = error.details.map(detail => ({
            field: detail.context.key,
            message: detail.message.replace(/['"]/g, '')
        }));
        
        console.log('Input validation failed:', validationErrors);
        
        return res.status(400).json({
            status: 'error',
            message: 'One or more required fields are missing or invalid.',
            errors: validationErrors
        });
    }

    req.body = value;

    next();
}

module.exports = { validateBotInput };