<script setup>
    // Vue utilities
    import { ref } from 'vue';

    // Material Web components
    import '@material/web/divider/divider.js';
    import '@material/web/textfield/outlined-text-field.js';
    import '@material/web/button/filled-button.js';
    import '@material/web/button/filled-tonal-button.js';
    import '@material/web/button/outlined-button.js';
    import '@material/web/icon/icon.js';
    import '@material/web/iconbutton/icon-button.js';
    import '@material/web/iconbutton/filled-tonal-icon-button.js';
    import '@material/web/ripple/ripple.js';

    const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;
    console.log('API base URL:', apiBaseUrl);

    const showPassword = ref(false);
    const loginDetails = ref({
        'email': '',
        'password': ''
    });
    const hasError = ref(false);
    const signup = ref(false);
    const errorMessage = ref('');
    const operationSuccess = ref(false);
    const successMessage = ref('');

    async function logUserIn() {
        console.log('Called logUserIn()');

        // Clear messages
        hasError.value = false;
        errorMessage.value = '';
        operationSuccess.value = false;
        successMessage.value = '';
        signup.value = false;

        if (loginDetails.value.email === '' || loginDetails.value.password === '') {
            console.warn('No email or password provided.');
            return;
        }

        try {
            const response = await fetch(`${apiBaseUrl}/api/login/`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    'email': loginDetails.value.email,
                    'password': loginDetails.value.password
                }),
                credentials: 'include'
            });

            console.log('Response:', response);

            const responseJson = await response.json();
            console.log('Response JSON from login:', responseJson);
            const responseText = responseJson.message;
            console.log('Response message:', responseText);

            if (!response.ok) {
                hasError.value = true;

                console.log('Response status from login:', response.status);
                if (response.status === 404) {
                    console.log('Response status 404.');
                    errorMessage.value = `${responseText}`;
                    signup.value = true;
                    return;
                }

                errorMessage.value = `${responseText}`;
                return;
            }

            console.log('User logged in successfully!');

            operationSuccess.value = true;
            successMessage.value = responseText;

            document.location.href = '/dashboard';
        } catch (error) {
            console.error('Error sending user details:', error);
            errorMessage.value = 'An unexpected error occurred, please try again later.';
        }
    }

    async function signUserUp() {
        // Clear messages
        hasError.value = false;
        errorMessage.value = '';
        operationSuccess.value = false;
        successMessage.value = '';
        signup.value = false;

        if (loginDetails.value.email === '' || loginDetails.value.password === '') {
            console.warn('No email or password provided.');
            return;
        }

        try {
            const response = await fetch(`${apiBaseUrl}/api/signup/`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    'email': loginDetails.value.email,
                    'password': loginDetails.value.password
                }),
                credentials: 'include'
            });

            console.log('Response:', response);

            const responseJson = await response.json();
            console.log('Response JSON from signup:', responseJson);
            const responseText = responseJson.message;
            console.log('Response message:', responseText);

            if (!response.ok) {
                hasError.value = true;

                console.log('Response status from signup:', response.status);

                errorMessage.value = `${responseText}`;
                return;
            }

            console.log('User signed up successfully!');

            operationSuccess.value = true;
            successMessage.value = responseText;
            
            logUserIn();
        } catch (error) {
            console.error('Error sending user details:', error);
            errorMessage.value = 'An unexpected error occurred, please try again later.';
        }
    }
</script>

<template>
    <div class="content-wrapper">
        <div class="login-box">
            <h1 class="login-text">Welcome!</h1>
            <p class="login-paragraph">Login to continue to Cordfriend AI.</p>
            <md-divider></md-divider>
            <div class="email-login-div">
                <p class="login-method-text">Login via an email address</p>
                <form class="email-login-form" @submit.prevent="logUserIn()">
                    <md-outlined-text-field type="email" label="Email address" supporting-text="Login or signup with your email address." required v-model="loginDetails.email" no-asterisk="true">
                    </md-outlined-text-field>
                    <md-outlined-text-field :type="showPassword ? 'text' : 'password'" label="Password" required v-model="loginDetails.password" no-asterisk="true">
                        <md-icon-button toggle slot="trailing-icon" @click="showPassword = !showPassword" type="button">
                            <md-icon>visibility</md-icon>
                            <md-icon slot="selected">visibility_off</md-icon>
                        </md-icon-button>
                    </md-outlined-text-field>
                    <md-filled-button type="submit" class="login-button">Login</md-filled-button>
                    <div class="error-div" :class="hasError ? 'has-error' : 'error-div'">
                        <p>{{ errorMessage }}</p>
                        <md-outlined-button v-show="signup" @click.stop="signUserUp()" type="button">Sign up instead</md-outlined-button>
                    </div>
                    <div class="success-div" :class="operationSuccess ? 'has-success' : 'success-div'">
                        <p>{{ successMessage }}</p>
                    </div>
                </form>
            </div>
            <md-divider></md-divider>
            <div class="oauth-login-div">
                <p>Login via an OAuth provider</p>
                <div class="oauth-providers">
                    <md-filled-tonal-button class="oauth-provider-button" disabled>
                        Continue with Google
                    </md-filled-tonal-button>
                    <md-filled-tonal-button class="oauth-provider-button" disabled>
                        Continue with Discord
                    </md-filled-tonal-button>
                </div>
            </div>
        </div>
        <div class="image-showcase">
        </div>
    </div>
</template>

<style scoped>
    .content-wrapper {
        width: 100vw;
        display: grid;
        grid-template-columns: 1fr 1fr;
        grid-template-rows: 1fr;
        height: 100vh;
    }

    .login-box {
        width: 100%;
        height: 100%;
        box-sizing: border-box;
        background-color: var(--md-sys-color-surface);
        color: var(--md-sys-color-on-surface);
        padding: 10px;
    }

    .image-showcase {
        display: flex;
        flex-direction: column;
        align-items: center;
        width: 100%;
        height: 100%;
        box-sizing: border-box;
        padding: 10px;
        background-color: var(--md-sys-color-primary);
    }

    .login-text {
        text-align: center;
        font-size: 75px;
        margin: 10px 0;
        color: var(--md-sys-color-primary);
    }

    .email-login-div, .oauth-login-div {
        text-align: center;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 10px;
        margin: 20px 0;
    }

    .login-paragraph {
        text-align: center;
        font-size: 30px;
    }

    md-outlined-text-field {
        width: 300px;
    }

    .login-button {
        width: 300px;
    }

    .email-login-form {
        text-align: center;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 10px;
    }

    .login-method-text {
        margin: 5px 0;
    }

    .signin-with-google {
        all: unset;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        width: 200px;
    }

    .signin-with-google-image {
        width: 200px;
        height: 50px;
    }

    .signin-with-discord {
        all: unset;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        background-color: var(--discord-blurple);
        color: var(--discord-light-blurple);
        border-radius: 100px;
        width: 200px;
        height: 45px;
        box-sizing: border-box;
        border: 1px solid #000000;
    }

    .oauth-providers {
        display: flex;
        flex-direction: column;
        gap: 10px;
        align-items: center;
        justify-content: center;
        padding: 10px;
        border-radius: 25px;
        background-color: var(--md-sys-color-inverse-primary);
        width: 300px;
        box-sizing: border-box;
    }

    .oauth-provider-button {
        width: 100%;
    }

    .error-div, .success-div {
        display: none;
        align-items: center;
        justify-content: center;
        border-radius: 25px;
        box-sizing: border-box;
        width: 300px;
        text-align: center;
    }

    .error-div.has-error, .success-div.has-success {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
    }

    .error-div {
        color: var(--md-sys-color-error);
    }

    .success-div {
        color: var(--md-sys-color-on-surface-variant);
    }

    .sign-up-link {
        color: var(--md-sys-color-primary);
        text-decoration: underline;
        cursor: pointer;
    }
</style>