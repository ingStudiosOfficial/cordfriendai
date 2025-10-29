const bcrypt = require('bcrypt');

function validateUser(passwordToValidate, passwordFromDatabase) {
    return new Promise((resolve, reject) => {
        bcrypt.compare(passwordToValidate, passwordFromDatabase, (err, result) => {
            if (err) {
                return reject(err);
            }

            if (result) {
                resolve(true);
            } else {
                resolve(false);
            }
        });
    });
}

module.exports = { validateUser };