<script setup>
	// Vue utilities
	import { RouterView, RouterLink } from 'vue-router';

	// Material Web components
	import '@material/web/iconbutton/outlined-icon-button.js';
	import '@material/web/icon/icon.js';

	// UI store
	import { useUiStore } from './stores/ui';

	// Utilities
	import logout from './utilities/logout';

	const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;
	console.log('API base URL:', apiBaseUrl);

	const uiStore = useUiStore();

	const triggerOpenAccountPanel = () => {
		uiStore.triggerAccountPanel();
	};
</script>

<template>
	<header v-if="!$route.meta.hideHeader">
		<div class="header-left-wrapper">
			<img src="/cordfriend_svg.svg" class="header-logo" />
			<h1>Cordfriend AI</h1>
		</div>
		<div class="header-buttons">
			<md-outlined-icon-button class="header-button" v-if="$route.meta.showAccountButton" @click="triggerOpenAccountPanel()">
				<md-icon>account_circle</md-icon>
			</md-outlined-icon-button>
			<md-outlined-icon-button class="header-button" v-if="$route.meta.showAccountButton" @click="logout(apiBaseUrl)">
				<md-icon>logout</md-icon>
			</md-outlined-icon-button>
		</div>
	</header>
	<main>
		<RouterView />
	</main>
</template>

<style scoped>

</style>