const bcrypt = require('bcrypt');

function validateUser(passwordToValidate, passwordFromDatabase) {
    return new Promise((resolve, reject) => {
        bcrypt.compare(passwordToValidate, passwordFromDatabase, (err, result) => {
            if (err) {
                console.error('Error while proccessing password:', err);
                return reject(err);
            }

            if (result) {
                console.log('Password correct.');
                resolve(true);
            } else {
                console.log('Password incorrect.');
                resolve(false);
            }
        });
    });
}

module.exports = { validateUser };