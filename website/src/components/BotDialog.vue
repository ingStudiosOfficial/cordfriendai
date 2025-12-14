<script setup>
    import { ref, watch } from 'vue';

    // Material Web components
    import '@material/web/textfield/outlined-text-field.js';
    import '@material/web/button/filled-button.js';
    import '@material/web/button/outlined-button.js';
    import '@material/web/icon/icon.js';
    import '@material/web/iconbutton/icon-button.js';
    import '@material/web/ripple/ripple.js';
    import '@material/web/focus/md-focus-ring.js';

    import { vibrate } from '@/utilities/vibrate';

    const props = defineProps({
        showBotDialog: Boolean,
        isEditingBot: Boolean,
        botToDisplay: Object
    });

    const emit = defineEmits(['closeBotDialog', 'refreshBots']);

    const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;
    const showApi = ref(false);
    const showWeatherApi = ref(false);
    const pendingImageFile = ref(null);
    const botProfilePicker = ref(null);
    const errorToDisplay = ref(null);
    let oldImageId = null;

    // Watch for when bot dialog opens with edit mode
    watch(() => props.botToDisplay, (newBot) => {
        if (newBot && props.isEditingBot) {
            oldImageId = newBot.image_id;
        }
    }, { immediate: true });

    function handleImageUpload(event) {
        pendingImageFile.value = event.target.files[0];
    }

    async function uploadImage() {
        if (!pendingImageFile.value) {
            return null;
        }

        const formData = new FormData();
        formData.append('bot-profile-picture', pendingImageFile.value);

        try {
            const response = await fetch(`${apiBaseUrl}/api/bot/image-upload/`, {
                method: 'POST',
                body: formData,
                credentials: 'include'
            });

            if (response.ok) {
                const data = await response.json();
                console.log('Image uploaded:', data);
                return data;
            } else {
                errorToDisplay.value = 'An unexpected error occurred while uploading image, please try again later.';
                throw new Error('Image upload failed.');
            }
        } catch (error) {
            console.error('Error uploading image:', error);
            errorToDisplay.value = 'An unexpected error occurred, please try again later.';
            return null;
        }
    }

    async function createBot() {
        vibrate([10]);

        const imageData = await uploadImage();
        if (imageData) {
            props.botToDisplay['image_id'] = imageData.fileId;
            props.botToDisplay['image_filename'] = imageData.filename;
        }

        console.log('Bot to display:', props.botToDisplay);
        const botToCreate = JSON.stringify(props.botToDisplay);

        try {
            const response = await fetch(`${apiBaseUrl}/api/bot/create/`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: botToCreate,
                credentials: 'include'
            });

            const responseJson = await response.json();
            console.log('Response JSON from creating bot:', responseJson);

            if (!response.ok) {
                console.error('Error while creating bot:', response.status);
                errorToDisplay.value = responseJson.message;
                return;
            }

            console.log('Bot created successfully!');

            if (!props.isEditingBot) {
                window.open("https://discord.com/oauth2/authorize?client_id=1425271680078581832&permissions=201575424&integration_type=0&scope=bot", '_blank');
            }

            closeBotDialog();
            emit('refreshBots');
        } catch (error) {
            console.error('Error while creating bot:', error);
            errorToDisplay.value = 'An unexpected error occurred, please try again later.';
        }
    }

    async function saveBot() {
        vibrate([10]);

        if (pendingImageFile.value) {
            const imageData = await uploadImage();
            if (imageData) {
                props.botToDisplay['image_id'] = imageData.fileId;
                props.botToDisplay['image_filename'] = imageData.filename;
            }
        }

        if (oldImageId) {
            props.botToDisplay['old_image_id'] = oldImageId;
        }

        const botToSave = JSON.stringify(props.botToDisplay);
        console.log('Saving bot:', botToSave);

        try {
            const response = await fetch(`${apiBaseUrl}/api/bot/edit/${props.botToDisplay._id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: botToSave,
                credentials: 'include'
            });

            console.log('Response status:', response.status);

            const responseJson = await response.json();
            console.log('Response JSON from saving bot:', responseJson);
            
            if (response.ok) {
                closeBotDialog();
                emit('refreshBots');
            } else {
                console.error('Error while saving bot:', responseJson.message);
                errorToDisplay.value = responseJson.message;
            }
        } catch (error) {
            console.error('Error while fetching response:', error);
            errorToDisplay.value = 'An unexpected error occurred, please try again later.';
        }
    }

    async function deleteBot() {
        vibrate([10]);

        const botToDelete = JSON.stringify(props.botToDisplay);
        console.log('Deleting bot:', botToDelete);

        if (!confirm(`Do you wish to delete the bot ${props.botToDisplay.name}? This action cannot be undone.`)) {
            console.log('User cancelled deletion.');
            return;
        }

        try {
            const response = await fetch(`${apiBaseUrl}/api/bot/delete/`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: botToDelete,
                credentials: 'include'
            });

            const responseJson = await response.json();
            console.log('Response JSON:', responseJson);

            if (response.ok) {
                closeBotDialog();
                emit('refreshBots');
            } else {
                errorToDisplay.value = responseJson.message;
            }
        } catch (error) {
            console.log('An error occurred while deleting bot:', error);
        }
    }

    async function deleteConv() {
        vibrate([10]);

        const botToDelete = JSON.stringify(props.botToDisplay);
        console.log('Deleting bot:', botToDelete);

        if (!confirm(`Do you wish to delete all conversations from ${props.botToDisplay.name}? This action cannot be undone.`)) {
            console.log('User cancelled deletion.');
            return;
        }

        try {
            const response = await fetch(`${apiBaseUrl}/api/bot/delete-conv/`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: botToDelete,
                credentials: 'include'
            });

            const responseJson = await response.json();
            console.log('Response JSON:', responseJson);

            if (response.ok) {
                closeBotDialog();
                emit('refreshBots');
            } else {
                errorToDisplay.value = responseJson.message;
            }
        } catch (error) {
            console.log('An error occurred while deleting conversations:', error);
        }
    }

    function openFilePicker() {
        vibrate([10]);

        botProfilePicker.value.click();
    }

    function closeBotDialog() {
        vibrate([10]);
        errorToDisplay.value = null;
        pendingImageFile.value = null;
        emit('closeBotDialog');
    }

    function truncateString(str, maxLength) {
        if (str.length > maxLength) {
            return str.slice(0, maxLength - 3) + "..."; 
        } else {
            return str;
        }
    }
</script>

<template>
    <div class="bot-dialog-backdrop" v-show="showBotDialog" v-if="botToDisplay">
        <form class="bot-dialog" @submit.prevent="isEditingBot ? saveBot() : createBot()">
            <h1 class="bot-dialog-header">Bot Settings</h1>
            <h2 class="bot-dialog-subheader">General</h2>
            <div class="bot-dialog-image-wrapper" v-if="botToDisplay.image_id">
                <img :src="`${apiBaseUrl}/api/bot/image-download/${botToDisplay.image_id}`" class="bot-image" />
            </div>
            <md-outlined-text-field 
                v-model="botToDisplay._id" 
                readOnly 
                class="dialog-settings-field" 
                label="Bot ID" 
                :required="isEditingBot" 
                no-asterisk="true" 
                supporting-text="The ID of the bot.">
            </md-outlined-text-field>
            <md-outlined-text-field 
                class="dialog-settings-field" 
                v-model="botToDisplay.name" 
                label="Bot name" 
                required 
                no-asterisk="true" 
                supporting-text="Your bot's display name.">
            </md-outlined-text-field>
            <md-outlined-text-field 
                class="dialog-settings-field" 
                v-model="botToDisplay.persona" 
                label="Bot persona" 
                required 
                no-asterisk="true" 
                supporting-text="Your bot's persona or instructions."
                type="textarea">
            </md-outlined-text-field>
            <div class="pfp-input">
                <p>Your bot's profile picture</p>
                <label class="file-upload-button" tabindex="0" @click="openFilePicker()" @keyup.enter="openFilePicker()" @keyup.space="openFilePicker()">
                    <md-ripple></md-ripple>
                    <md-focus-ring style="--md-focus-ring-shape: 25px"></md-focus-ring>
                    <md-icon>upload</md-icon>
                </label>
                <input 
                    type="file" 
                    ref="botProfilePicker" 
                    name="bot-profile-picture" 
                    accept="image/*" 
                    @change="handleImageUpload($event)" 
                    :required="!isEditingBot" 
                    style="display: none;">
                <p v-if="pendingImageFile" class="file-chosen">{{ pendingImageFile.name }}</p>
                <p v-else-if="botToDisplay.image_filename" class="file-chosen">{{ botToDisplay.image_filename }}</p>
                <p v-else class="file-chosen">No file chosen</p>
            </div>
            <h2 class="bot-dialog-subheader">Conversations</h2>
            <div class="conv-div" v-show="isEditingBot">
                <div class="conv-group" v-if="botToDisplay.conversations && botToDisplay.conversations.length !== 0" v-for="conversation_group in botToDisplay.conversations.slice(0, 3)">
                    <div v-if="conversation_group.user" class="message-bubble right"><b>{{ conversation_group.user.name + ':' }}</b> {{ truncateString(conversation_group.user.message, 200) }}</div>
                    <div v-if="conversation_group.bot" class="message-bubble left"><b>{{ botToDisplay.name }}:</b> {{ truncateString(conversation_group.bot, 200) }}</div>
                </div>
                <p v-if="botToDisplay.conversations && botToDisplay.conversations.length === 0">No conversations yet.</p>
            </div>
            <h2 class="bot-dialog-subheader">Server</h2>
            <md-outlined-text-field 
                class="dialog-settings-field" 
                v-model="botToDisplay.server_id" 
                label="Server ID" 
                required 
                no-asterisk="true" 
                supporting-text="The Discord server ID of the bot.">
            </md-outlined-text-field>
            <h2 class="bot-dialog-subheader">Administrator</h2>
            <md-outlined-text-field 
                class="dialog-settings-field" 
                v-model="botToDisplay.user_id" 
                label="User ID" 
                required 
                no-asterisk="true" 
                supporting-text="Your Discord user ID.">
            </md-outlined-text-field>
            <h2 class="bot-dialog-subheader">Google AI</h2>
            <md-outlined-text-field 
                class="dialog-settings-field" 
                v-model="botToDisplay.google_ai_api" 
                label="Google AI API key" 
                required 
                no-asterisk="true" 
                supporting-text="Your Google AI API key." 
                :type="showApi ? 'text' : 'password'">
                <md-icon-button toggle slot="trailing-icon" @click="showApi = !showApi" type="button">
                    <md-icon>visibility</md-icon>
                    <md-icon slot="selected">visibility_off</md-icon>
                </md-icon-button>
            </md-outlined-text-field>
            <h2 class="bot-dialog-subheader">OpenWeatherMap</h2>
            <md-outlined-text-field 
                class="dialog-settings-field" 
                v-model="botToDisplay.openweathermap_api" 
                label="OpenWeatherMap API key" 
                required 
                no-asterisk="true" 
                supporting-text="Your OpenWeatherMap API key." 
                :type="showWeatherApi ? 'text' : 'password'">
                <md-icon-button toggle slot="trailing-icon" @click="showWeatherApi = !showWeatherApi" type="button">
                    <md-icon>visibility</md-icon>
                    <md-icon slot="selected">visibility_off</md-icon>
                </md-icon-button>
            </md-outlined-text-field>
            <div class="danger-zone" v-show="isEditingBot">
                <h2 class="bot-dialog-subheader">Danger Zone</h2>
                <p>This action cannot be undone. All conversations associated with the bot will be lost</p>
                <md-outlined-button class="delete-button" type="button" @click="deleteConv()">Delete conversations</md-outlined-button>
                <p>This action cannot be undone. All data associated with the bot will be lost.</p>
                <md-outlined-button class="delete-button" type="button" @click="deleteBot()">Delete bot</md-outlined-button>
            </div>
            <div class="error-div">
                <p>{{ errorToDisplay }}</p>
            </div>
            <div class="dialog-actions-div">
                <md-outlined-button type="button" @click="closeBotDialog()">Cancel</md-outlined-button>
                <md-filled-button type="submit">Save</md-filled-button>
            </div>
        </form>
    </div>
</template>

<style scoped>
    .bot-dialog-backdrop {
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

    .bot-dialog {
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

    .bot-dialog * {
        margin: 0;
    }

    .bot-dialog-header, .bot-dialog-subheader {
        color: var(--md-sys-color-primary);
    }

    .bot-dialog-image-wrapper {
        display: flex;
        align-items: center;
        justify-content: center;
        overflow: hidden;
        border-radius: 50%;
        width: 150px;
        height: 150px;
        flex-shrink: 0;
    }

    .bot-image {
        width: 100%;
        height: 100%;
        object-fit: cover;
    }

    .dialog-settings-field {
        width: 50%;
        color: var(--md-sys-color-on-primary-container);
    }

    .pfp-input {
        display: flex;
        flex-direction: column;
        justify-content: center;
        background-color: var(--md-sys-color-surface);
        color: var(--md-sys-color-on-surface);
        border-radius: 25px;
        padding: 20px;
        gap: 10px;
        width: 50%;
        box-sizing: border-box;
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

    .file-upload-button {
        display: block;
        position: relative;
        background-color: var(--md-sys-color-primary-container);
        color: var(--md-sys-color-on-primary-container);
        padding: 10px;
        border-radius: 25px;
        cursor: pointer;
        display: inline-block;
        box-sizing: border-box;
        text-align: center;
        height: 50px;
        outline: none;
    }

    .file-chosen {
        word-wrap: break-word;
    }

    .error-div {
        color: var(--md-sys-color-error);
    }

    .conv-div {
        display: flex;
        flex-direction: column;
        justify-content: center;
        background-color: var(--md-sys-color-surface);
        color: var(--md-sys-color-on-surface);
        border-radius: 25px;
        padding: 20px;
        width: 50%;
        box-sizing: border-box;
    }

    .message-bubble {
        position: relative;
        background-color: var(--md-sys-color-surface-variant);
        color: var(--md-sys-color-on-surface-variant);
        border-radius: 10px;
        padding: 10px 15px;
        margin: 10px;
        max-width: 70%;
        word-wrap: break-word;
    }

    .message-bubble.left::before {
        content: '';
        position: absolute;
        top: 50%;
        left: -10px;
        transform: translateY(-50%);
        border-width: 10px 10px 10px 0;
        border-style: solid;
        border-color: transparent var(--md-sys-color-surface-variant) transparent transparent;
    }

    .message-bubble.right {
        background-color: var(--md-sys-color-primary-container);
        color: var(--md-sys-color-on-primary-container);
        margin-left: auto;
    }

    .message-bubble.right::before {
        content: '';
        position: absolute;
        top: 50%;
        right: -10px;
        transform: translateY(-50%);
        border-width: 10px 0 10px 10px;
        border-style: solid;
        border-color: transparent transparent transparent var(--md-sys-color-primary-container);
    }

    @media (max-width: 768px) {
        .bot-dialog {
            width: 90%;
            height: 90%;
            border-radius: 25px;
        }

        .dialog-settings-field, .pfp-input, .danger-zone, .conv-div {
            width: 80%;
        }
    }
</style>