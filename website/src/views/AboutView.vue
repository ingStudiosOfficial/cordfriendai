<script setup>
    import '@material/web/divider/divider.js';
    import '@material/web/button/filled-button.js';
</script>

<template>
    <div class="content-wrapper">
        <div class="hero-section">
            <img class="hero-logo" src="/cordfriend_svg.svg" />
            <h1 class="cordfriend-title">Cordfriend AI</h1>
            <h1 class="hero-text">Your <span class="hero-emphasis">Free</span>, <span class="hero-emphasis">Open-source</span> AI Discord Bot</h1>
            <md-filled-button href="/login" class="get-started">Get started for free</md-filled-button>
        </div>
        <md-divider></md-divider>
        <h1 class="section-title">Why Cordfriend AI?</h1>
        <ul class="why-cordfriend">
            <li>Free</li>
            <li>Open-source</li>
            <li>Fast</li>
        </ul>
        <md-divider></md-divider>
        <h1 class="section-title">Pricing</h1>
        <p class="section-para"><span class="hero-emphasis">Completely free</span>, <span class="hero-emphasis">no subscriptions</span>, <span class="hero-emphasis">no hidden costs</span>. Just <span class="hero-emphasis">BYOK</span>!</p>
        <md-divider></md-divider>
        <h1 class="section-title">Built with DiscordGo</h1>
        <p class="section-para">That's right - Cordfriend is built with the <span class="hero-emphasis">DiscordGo</span> framework in <span class="hero-emphasis">Golang</span>.</p>
        <p class="section-para">Why did I build it in Golang? Because all other bots are made in JavaScript and Python.</p>
        <md-divider></md-divider>
        <h1 class="section-title">What Are You Waiting For?</h1>
        <p class="section-para">Sign up today by clicking the button below!</p>
        <md-filled-button href="/login" class="get-started">Get started</md-filled-button>
        <md-divider></md-divider>
    </div>
</template>

<style scoped>
    .content-wrapper {
        display: flex;
        flex-direction: column;
        align-items: center;
        text-align: center;
        margin-bottom: 50px;
    }

    .hero-logo {
        width: 200px;
    }

    .cordfriend-title {
        font-size: 80px;
        color: var(--md-sys-color-primary);
    }

    .hero-section {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100vh;
    }

    .hero-text {
        font-size: 50px;
    }

    .hero-emphasis {
        padding: 5px;
        border-radius: 10px;
        background-color: var(--md-sys-color-primary-container);
        color: var(--md-sys-color-on-primary-container);
    }

    .section-title {
        color: var(--md-sys-color-primary);
        font-size: 40px;
    }

    .why-cordfriend {
        font-size: 30px;
        list-style-type: none;
    }

    .why-cordfriend li {
        padding: 5px;
        border-radius: 10px;
        background-color: var(--md-sys-color-primary-container);
        color: var(--md-sys-color-on-primary-container);
        margin: 10px 0;
    }

    .section-para {
        font-size: 30px;
    }

    .get-started {
        margin: 10px;
    }
</style>