import { defineStore } from 'pinia'

export const useUiStore = defineStore('ui', {
    state: () => ({
        openAccountPanel: false
    }),
    actions: {
        triggerAccountPanel() {
            this.openAccountPanel = true
        },
        resetAccountPanel() {
            this.openAccountPanel = false
        }
    }
})