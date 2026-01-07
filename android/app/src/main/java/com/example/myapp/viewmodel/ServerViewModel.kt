package com.example.myapp.viewmodel

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.myapp.network.ApiClient
import com.example.myapp.network.AuthService
import kotlinx.coroutines.launch

sealed class ServerState {
    object Idle : ServerState()
    object Loading : ServerState()
    object Success : ServerState()
    data class Error(val message: String) : ServerState()
}

class ServerViewModel : ViewModel() {

    private val _state = MutableLiveData<ServerState>(ServerState.Idle)
    val state: LiveData<ServerState> = _state

    fun validateServer(baseUrl: String) {
        viewModelScope.launch {
            _state.value = ServerState.Loading

            try {
                // Ensure baseUrl ends with /
                val formattedUrl = if (baseUrl.endsWith("/")) baseUrl else "$baseUrl/"

                val authService = ApiClient.createClient(formattedUrl).create(AuthService::class.java)
                val response = authService.health()

                if (response.isSuccessful && response.body()?.status == "success" &&
                    response.body()?.body?.status == "healthy") {
                    _state.value = ServerState.Success
                } else {
                    _state.value = ServerState.Error("Server validation failed")
                }
            } catch (e: Exception) {
                _state.value = ServerState.Error(e.message ?: "Unknown error occurred")
            }
        }
    }
}
