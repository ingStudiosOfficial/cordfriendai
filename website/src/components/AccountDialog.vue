<script setup>
    import { ref, watch } from 'vue';

    // Material Web components
    import '@material/web/textfield/outlined-text-field.js';
    import '@material/web/button/filled-button.js';
    import '@material/web/button/outlined-button.js';
    import '@material/web/icon/icon.js';
    import '@material/web/iconbutton/icon-button.js';

    import { vibrate } from '@/utilities/vibrate';

    const props = defineProps({
        showAccountDialog: Boolean,
        userAccount: Object
    });

    const emit = defineEmits(['closeAccountDialog']);

    const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;
    const showOldPassword = ref(false);
    const showNewPassword = ref(false);
    const errorToDisplay = ref(null);
    const accountToEdit = ref(props.userAccount);

    async function editAccount() {
        vibrate([10]);

        errorToDisplay.value = null;

        const dataToSend = JSON.stringify(accountToEdit.value);
        console.log('Editing account:', dataToSend);

        try {
            const response = await fetch(`${apiBaseUrl}/api/user/edit/`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: dataToSend,
                credentials: 'include'
            });

            console.log('Response status from editing the user:', response.status);

            const responseJson = await response.json();
            console.log('Response JSON from editing the user:', responseJson);

            const responseMessage = responseJson.message;
            console.log('Response message from editing the user:', responseMessage);

            if (response.ok) {
                console.log('User updated successfully.');
                closeAccountDialog();
            } else {
                console.error('An error occurred while editing user.');
                errorToDisplay.value = responseMessage;
            }
        } catch (error) {
            console.error('An error occurred while editing user:', error);
            errorToDisplay.value = 'An unexpected error occurred, please try again later.';
        }
    }

    async function deleteAccount() {
        vibrate([10]);

        if (!confirm('Do you wish to delete your account and all your bots? This action cannot be undone.')) {
            console.log('User canceled deletion of account.');
            return;
        }

        errorToDisplay.value = null;

        try {
            const response = await fetch(`${apiBaseUrl}/api/user/delete/`, {
                method: 'DELETE',
                credentials: 'include'
            });

            console.log('Response status from deleting user:', response.status);
            
            const responseJson = await response.json();
            console.log('JSON from deleting user:', responseJson);

            const responseMessage = responseJson.message;
            console.log('Message from deleting user:', responseMessage);

            if (response.ok) {
                closeAccountDialog();
                window.location.href = '/login';
            } else {
                errorToDisplay.value = responseMessage;
            }
        } catch (error) {
            console.error('Error while deleting user:', error);
            errorToDisplay.value = 'An unexpected error occurred, please try again later.';
        }
    }

    function closeAccountDialog() {
        vibrate([10]);

        errorToDisplay.value = null;
        showOldPassword.value = false;
        showNewPassword.value = false;
        emit('closeAccountDialog');
    }

    watch(
        () => props.userAccount,
        (newUserAccount) => {
            accountToEdit.value = newUserAccount ? { ...newUserAccount } : null; 
        },
        { deep: true, immediate: true }
    );
</script>

<template>
    <div class="account-dialog-backdrop" v-show="showAccountDialog" v-if="accountToEdit">
        <form class="account-dialog" @submit.prevent="editAccount()">
            <h1 class="account-dialog-header">Account Settings</h1>
            
            <h2 class="account-dialog-subheader">General</h2>
            <md-outlined-text-field 
                v-model="accountToEdit.id" 
                readOnly 
                class="dialog-settings-field" 
                label="Account ID" 
                no-asterisk="true" 
                supporting-text="The ID of your account.">
            </md-outlined-text-field>
            <md-outlined-text-field 
                v-model="accountToEdit.email" 
                class="dialog-settings-field" 
                label="Email" 
                no-asterisk="true" 
                supporting-text="The email associated with your account.">
            </md-outlined-text-field>
            
            <h2 class="account-dialog-subheader">Change Password</h2>
            <md-outlined-text-field 
                class="dialog-settings-field" 
                v-model="accountToEdit.old_password" 
                label="Old password" 
                no-asterisk="true" 
                supporting-text="Your current password associated with your account." 
                :type="showOldPassword ? 'text' : 'password'">
                <md-icon-button toggle slot="trailing-icon" @click="showOldPassword = !showOldPassword" type="button">
                    <md-icon>visibility</md-icon>
                    <md-icon slot="selected">visibility_off</md-icon>
                </md-icon-button>
            </md-outlined-text-field>
            <md-outlined-text-field 
                class="dialog-settings-field" 
                v-model="accountToEdit.new_password" 
                label="New password" 
                no-asterisk="true" 
                supporting-text="Your password you wish to set for your account." 
                :type="showNewPassword ? 'text' : 'password'">
                <md-icon-button toggle slot="trailing-icon" @click="showNewPassword = !showNewPassword" type="button">
                    <md-icon>visibility</md-icon>
                    <md-icon slot="selected">visibility_off</md-icon>
                </md-icon-button>
            </md-outlined-text-field>
            
            <div class="danger-zone">
                <h2 class="account-dialog-subheader">Danger Zone</h2>
                <p>This action cannot be undone. All data associated with your account will be lost.</p>
                <md-outlined-button class="delete-button" type="button" @click="deleteAccount()">
                    Delete account
                </md-outlined-button>
            </div>

            <div class="error-div" v-show="errorToDisplay">
                <p>{{ errorToDisplay }}</p>
            </div>
            
            <div class="dialog-actions-div">
                <md-outlined-button type="button" @click="closeAccountDialog()">Cancel</md-outlined-button>
                <md-filled-button type="submit">Save</md-filled-button>
            </div>
        </form>
    </div>
</template>

<style scoped>
    .account-dialog-backdrop {
        position: fixed;
        top: 0;
        left: 0;
        width: 100vw;
        height: 100vh;
        background-color: rgba(0, 0, 0, 0.5);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
    }

    .account-dialog {
        width: min(600px, 90vw);
        height: min(600px, 90vh);
        background-color: var(--md-sys-color-primary-container);
        padding: 20px;
        border-radius: 25px;
        display: flex;
        flex-direction: column;
        align-items: center;
        overflow-y: scroll;
        gap: 20px;
        box-sizing: border-box;
        color: var(--md-sys-color-on-primary-container);
    }

    .account-dialog * {
        margin: 0;
    }

    .account-dialog-header, .account-dialog-subheader {
        color: var(--md-sys-color-primary);
    }

    .dialog-settings-field {
        width: 50%;
        color: var(--md-sys-color-on-primary-container);
    }

    .dialog-actions-div {
        display: flex;
        flex-direction: row;
        align-items: center;
        justify-content: center;
        gap: 10px;
    }

    .danger-zone {
        display: flex;
        flex-direction: column;
        gap: 20px;
        align-items: center;
        justify-content: center;
        width: 50%;
        border-radius: 25px;
        border: 2px solid var(--md-sys-color-error);
        padding: 20px;
        box-sizing: border-box;
    }

    .danger-zone * {
        color: var(--md-sys-color-error);
        text-align: center;
    }

    .delete-button {
        --md-sys-color-outline: var(--md-sys-color-error);
        --md-sys-color-primary: var(--md-sys-color-error);
    }

    .error-div {
        color: var(--md-sys-color-error);
    }

    @media (max-width: 768px) {
        .account-dialog {
            width: 90%;
            height: 90%;
            border-radius: 25px;
        }

        .dialog-settings-field, .danger-zone {
            width: 80%;
        }
    }
</style>