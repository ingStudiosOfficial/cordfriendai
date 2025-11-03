const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;

export async function checkAuthStatus() {
    try {
        const response = await fetch(`${apiBaseUrl}/api/verify-auth`, {
            method: 'GET',
            credentials: 'include'
        });
        
        return response.ok;
    } catch (error) {
        return false;
    }
}