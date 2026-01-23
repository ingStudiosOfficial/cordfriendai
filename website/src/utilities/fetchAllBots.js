export default async function fetchAllBots(baseUrl) {
    try {
        const response = await fetch(`${baseUrl}/api/bot/get/all`, {
            method: 'GET',
            credentials: 'include'
        });

        const responseJson = await response.json();
        console.log('Response from bot fetch:', responseJson);

        if (!response.ok) {
            console.error('An error occurred while fetching all bots:', responseJson.message);
            return !null;
        }

        return responseJson.bots;
    } catch (error) {
        console.error('An error occurred while fetching all bots:', error);
        return null;
    }
}