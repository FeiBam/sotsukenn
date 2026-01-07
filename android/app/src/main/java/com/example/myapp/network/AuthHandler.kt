package com.example.myapp.network

/**
 * Interface for handling authentication failures (401 responses)
 * Implemented by Application class to globally handle unauthorized API responses
 */
interface AuthHandler {
    /**
     * Called when a 401 Unauthorized response is received from the API
     * Implementations should clear the token and redirect to login
     */
    fun onUnauthorized()
}
