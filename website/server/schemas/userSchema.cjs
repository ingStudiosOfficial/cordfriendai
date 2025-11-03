const Joi = require('joi');

const userUpdateSchema = Joi.object({
    email: Joi.string().email().required(),
    old_password: Joi.string().optional().allow(''), 
    new_password: Joi.string().optional().allow(''),
})
.and('old_password', 'new_password')
.messages({
    'object.and': 'To change your password, you must provide both the old and new passwords.'
});

module.exports = userUpdateSchema;