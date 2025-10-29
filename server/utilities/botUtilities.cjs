const { ObjectId } = require('mongodb');

async function deleteBot(botsCollection, usersCollection, botImagesBucket, botId, userId, imageId, removeFromUser = true) {
    // Validate all IDs upfront
    if (!ObjectId.isValid(botId)) {
        throw new Error('Invalid bot ID format.');
    }

    if (!ObjectId.isValid(userId)) {
        throw new Error('Invalid user ID format.');
    }

    if (!ObjectId.isValid(imageId)) {
        throw new Error('Invalid bot image ID format.');
    }

    const botObjectId = new ObjectId(botId);
    const botImageId = new ObjectId(imageId);
    const userObjectId = new ObjectId(userId);

    // Delete from bots collection
    const deleteResult = await botsCollection.deleteOne({ '_id': botObjectId });
    if (deleteResult.deletedCount === 0) {
        throw new Error('Bot not found.');
    }

    // Remove from user's bots array
    if (removeFromUser === true) {
        const updateResult = await usersCollection.updateOne(
            { '_id': userObjectId },
            { $pull: { 'bots': botObjectId } }
        );

        if (updateResult.matchedCount === 0) {
            throw new Error('User not found.');
        }
    }

    // Delete image from GridFS
    await botImagesBucket.delete(botImageId);

    return { message: 'Bot deleted successfully.' };
}

module.exports = { deleteBot };