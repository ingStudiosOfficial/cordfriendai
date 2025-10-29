export default async function logout(apiBaseUrl) {
    if (!confirm('Are you sure you want to logout?')) {
        console.log('User cancelled logout.');
        return;
    }

    try {
        const response = await fetch(`${apiBaseUrl}/api/logout/`, {
            method: 'POST',
            credentials: 'include'
        });

        console.log('Response status from logout:', response.status);

        const responseJson = await response.json();
        console.log('Response JSON from logout:', responseJson);

        if (response.ok) {
            document.location.href = '/login';
        }
    } catch (error) {
        console.error('Error while logging user out:', error);
    }
}