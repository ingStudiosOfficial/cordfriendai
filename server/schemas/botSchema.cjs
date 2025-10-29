const Joi = require('joi');

const botUpdateSchema = Joi.object({
    id: Joi.string().length(24).hex().optional(),
    name: Joi.string().required(),
    server_id: Joi.string().required(),
    user_id: Joi.string().required(),
    google_ai_api: Joi.string().required(),
    image_id: Joi.string().length(24).hex().required(),
    image_filename: Joi.string().required(),
    old_image_id: Joi.string().length(24).hex().optional()
});

module.exports = botUpdateSchema;