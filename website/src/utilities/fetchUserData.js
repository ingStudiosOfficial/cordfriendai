export default async function fetchUserData(apiBaseUrl) {
    try {
        const response = await fetch(`${apiBaseUrl}/api/user/get/`, {
            method: 'GET',
            credentials: 'include'
        });

        console.log('Response status from user fetch:', response.status);

        const responseJson = await response.json();
        console.log('Response JSON from user fetch:', responseJson);
        
        const responseMessage = responseJson.message;
        console.log('Response message from user fetch:', responseMessage);

        if (response.ok) {
            return responseJson.user;
        } else {
            return null;
        }
    } catch (error) {
        console.error('An error occurred while fetching user:', error);
        return null;
    }
}