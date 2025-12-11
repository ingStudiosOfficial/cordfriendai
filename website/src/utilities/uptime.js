export function formatUptime(ms) {
    let totalSeconds = Math.floor(ms / 1000);

    const years = Math.floor(totalSeconds / (365 * 24 * 3600));
    totalSeconds %= 365 * 24 * 3600;

    const days = Math.floor(totalSeconds / (24 * 3600));
    totalSeconds %= 24 * 3600;

    const hours = Math.floor(totalSeconds / 3600);
    totalSeconds %= 3600;

    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;

    let parts = [];
    if (years > 0) parts.push(`${years}y`);
    if (days > 0) parts.push(`${days}d`);
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    parts.push(`${seconds}s`);

    return parts.join(" ");
}