const Joi = require('joi');

const userUpdateSchema = Joi.object({
    email: Joi.string().email().required(),
    old_password: Joi.string().optional(),
    new_password: Joi.string().optional()
}).and('old_password', 'new_password')
.messages({
    'object.and': 'You need both your old and new passwords to change your passqord.'
});

module.exports = userUpdateSchema;