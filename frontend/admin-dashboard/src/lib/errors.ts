/**
 * API Not Connected Error
 * 
 * Standard error for unimplemented backend endpoints.
 * UI CONTRACT MODE: All missing APIs throw this error.
 */

export class APINotConnectedError extends Error {
    constructor(endpoint: string) {
        super(`API_NOT_CONNECTED: ${endpoint}`);
        this.name = 'APINotConnectedError';
    }
}

/**
 * Check if error is API not connected
 */
export function isAPINotConnected(error: unknown): boolean {
    return error instanceof APINotConnectedError ||
        (error instanceof Error && error.message.includes('API_NOT_CONNECTED'));
}

/**
 * Get user-friendly message for API not connected error
 */
export function getAPINotConnectedMessage(error: unknown): string {
    if (isAPINotConnected(error)) {
        return 'This feature requires backend API integration. The UI is ready and serves as the contract for implementation.';
    }
    return error instanceof Error ? error.message : 'An unexpected error occurred';
}
