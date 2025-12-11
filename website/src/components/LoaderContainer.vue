<script setup>
const props = defineProps({
    loadingText: String,
    loaderColor: String,
});
</script>

<template>
    <div class="loader-wrapper">
        <span class="loader" :style="{
            'border-color': props.loaderColor,
            'border-bottom-color': 'transparent',
        }"></span>
        <p class="loader-text" :style="{
            'color': props.loaderColor,
        }">{{ props.loadingText }}</p>
    </div>
</template>

<style scoped>
.loader-wrapper {
    display: flex;
    flex-direction: column;
    align-items: center;
}

.loader {
    width: 48px;
    height: 48px;
    border-width: 5px;
    border-style: solid;
    
    border-color: #FFF; 
    
    border-bottom-color: transparent; 
    
    border-radius: 50%;
    display: inline-block;
    box-sizing: border-box;
    animation: rotation 1s linear infinite;
}

.loader-text {
    font-size: 1.5rem;
    text-align: center;
}

@keyframes rotation {
    0% {
        transform: rotate(0deg);
    }

    100% {
        transform: rotate(360deg);
    }
}
</style>