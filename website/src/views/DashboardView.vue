<script setup>
    import { ref, onMounted, watch, onUnmounted } from 'vue';
    import '@material/web/fab/fab.js';
    import '@material/web/focus/md-focus-ring.js';
    import '@material/web/icon/icon.js';
    import '@material/web/ripple/ripple.js';

    import fetchAllBots from '@/utilities/fetchAllBots';
    import fetchUserData from '@/utilities/fetchUserData';
    import { vibrate } from '@/utilities/vibrate';

    import { useUiStore } from '../stores/ui';

    import BotDialog from '@/components/BotDialog.vue';
    import AccountDialog from '@/components/AccountDialog.vue';

    const uiStore = useUiStore();
    const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;

    const showBotDialog = ref(false);
    const serverBots = ref(null);
    const botToDisplay = ref({});
    const isEditingBot = ref(false);
    const showAccountDialog = ref(false);
    const userAccount = ref(null);

    let fetchedUserAccount;

    // Bot Dialog handlers
    function showCreateBotDialog() {
        vibrate([10]);
        showAccountDialog.value = false;
        isEditingBot.value = false;
        botToDisplay.value = {};
        showBotDialog.value = true;
    }

    function showEditBotDialog(bot) {
        vibrate([10]);
        showAccountDialog.value = false;
        isEditingBot.value = true;
        botToDisplay.value = bot;
        showBotDialog.value = true;
    }

    function closeBotDialog() {
        vibrate([10]);
        botToDisplay.value = {};
        showBotDialog.value = false;
    }

    async function refreshBots() {
        serverBots.value = await fetchAllBots(apiBaseUrl);
    }

    watch(showBotDialog, (isOpen) => {
        const handleEscape = (e) => {
            if (e.key === 'Escape') {
                closeBotDialog();
            }
        };
        
        if (isOpen) {
            document.addEventListener('keydown', handleEscape);
        }
        
        return () => document.removeEventListener('keydown', handleEscape);
    });

    // Account Dialog handlers
    function openAccountDialog() {
        vibrate([10]);
        showBotDialog.value = false;
        showAccountDialog.value = true;
        userAccount.value = fetchedUserAccount;
    }

    function closeAccountDialog() {
        vibrate([10]);
        uiStore.resetAccountPanel();
        showAccountDialog.value = false;
        userAccount.value = null;
    }

    watch(showAccountDialog, (isOpen) => {
        const handleEscape = (e) => {
            if (e.key === 'Escape') {
                closeAccountDialog();
            }
        };
        
        if (isOpen) {
            document.addEventListener('keydown', handleEscape);
        }
        
        return () => document.removeEventListener('keydown', handleEscape);
    });

    watch(() => uiStore.openAccountPanel, (isTriggered) => {
        if (isTriggered) {
            openAccountDialog();
        }
    });

    onMounted(async () => {
        serverBots.value = await fetchAllBots(apiBaseUrl);
        fetchedUserAccount = await fetchUserData(apiBaseUrl);
    });

    onUnmounted(() => {
        if (uiStore.openAccountPanel) {
            uiStore.resetAccountPanel();
        }
    });
</script>

<template>
    <div class="content-wrapper">
        <div class="bot-cards">
            <button class="bot-card" v-for="serverBot in serverBots" :key="serverBot._id" @click="showEditBotDialog(serverBot)">
                <md-ripple></md-ripple>
                <md-focus-ring style="--md-focus-ring-shape: 25px"></md-focus-ring>
                <div class="bot-image-wrapper">
                    <img :src="`${apiBaseUrl}/api/bot/image-download/${serverBot.image_id}`" class="bot-image" />
                </div>
                <h1>{{ serverBot.name }}</h1>
            </button>
        </div>
    </div>

    <BotDialog 
        :show-bot-dialog="showBotDialog" 
        :is-editing-bot="isEditingBot" 
        :bot-to-display="botToDisplay" 
        @close-bot-dialog="closeBotDialog()"
        @refresh-bots="refreshBots()">
    </BotDialog>

    <AccountDialog
        :show-account-dialog="showAccountDialog"
        :user-account="userAccount"
        @close-account-dialog="closeAccountDialog()">
    </AccountDialog>

    <md-fab class="add-button" label="Create" @click="showCreateBotDialog()">
        <md-icon slot="icon">add</md-icon>
    </md-fab>
    <md-fab class="mobile-add-button" size="large" @click="showCreateBotDialog()">
        <md-icon slot="icon">add</md-icon>
    </md-fab>
</template>

<style scoped>
    .content-wrapper {
        display: flex;
        flex-direction: column;
        width: 100vw;
        height: auto;
        align-items: center;
        padding-top: 10px;
    }

    .bot-cards {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        grid-template-rows: auto;
        width: 90%;
        box-sizing: border-box;
        gap: 10px;
    }

    .bot-card {
        all: unset;
        position: relative;
        display: flex;
        flex-direction: column;
        align-items: center;
        cursor: pointer;
        width: 100%;
        background-color: var(--md-sys-color-primary-container);
        color: var(--md-sys-color-on-primary-container);
        box-sizing: border-box;
        padding: 10px;
        border-radius: 25px;
        text-align: center;
        user-select: none;
    }

    .bot-image {
        width: 100%;
        height: 100%;
        object-fit: cover;
    }

    .bot-image-wrapper {
        display: flex;
        align-items: center;
        justify-content: center;
        overflow: hidden;
        width: 50%;
        aspect-ratio: 1 / 1;
        border-radius: 50%;
    }

    .add-button {
        position: fixed;
        margin: 25px;
        bottom: 0;
        right: 0;
    }

    .mobile-add-button {
        display: none;
    }
    
    @media (max-width: 768px) {
        .bot-cards {
            display: flex;
            flex-direction: column;
            align-items: center;
        }

        .add-button {
            display: none;
        }

        .mobile-add-button {
            display: block;
            position: fixed;
            margin: 25px;
            bottom: 0;
            right: 0;
        }
    }
</style>