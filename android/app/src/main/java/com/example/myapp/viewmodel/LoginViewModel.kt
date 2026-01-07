package com.example.myapp.viewmodel

import android.util.Log
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.myapp.model.FcmTokenRequest
import com.example.myapp.model.LoginRequest
import com.example.myapp.network.ApiClient
import com.example.myapp.network.AuthService
import com.example.myapp.utils.DeviceInfoUtils
import com.example.myapp.utils.PreferenceManager
import kotlinx.coroutines.launch

sealed class LoginState {
    object Idle : LoginState()
    object Loading : LoginState()
    data class Success(val token: String, val username: String, val user: com.example.myapp.model.UserInfo?) : LoginState()
    data class Error(val message: String) : LoginState()
}

class LoginViewModel : ViewModel() {

    private val _state = MutableLiveData<LoginState>(LoginState.Idle)
    val state: LiveData<LoginState> = _state

    private var preferenceManager: PreferenceManager? = null

    companion object {
        private const val TAG = "LoginViewModel"
    }

    /**
     * 设置 PreferenceManager（由 Activity 调用）
     */
    fun setPreferenceManager(preferenceManager: PreferenceManager) {
        this.preferenceManager = preferenceManager
    }

    fun login(baseUrl: String, username: String, password: String) {
        viewModelScope.launch {
            _state.value = LoginState.Loading

            try {
                // Ensure baseUrl ends with /
                val formattedUrl = if (baseUrl.endsWith("/")) baseUrl else "$baseUrl/"

                val authService = ApiClient.createClient(formattedUrl).create(AuthService::class.java)
                val request = LoginRequest(username, password)
                val response = authService.login(request)

                if (response.isSuccessful) {
                    val responseBody = response.body()
                    if (responseBody?.status == "success" && responseBody.body.token.isNotEmpty()) {
                        val jwtToken = responseBody.body.token

                        // 登录成功后发送 FCM Token
                        submitFcmToken(formattedUrl, jwtToken)

                        _state.value = LoginState.Success(
                            jwtToken,
                            username,
                            responseBody.body.user
                        )
                    } else {
                        _state.value = LoginState.Error(responseBody?.message ?: "Login failed: Invalid response")
                    }
                } else {
                    _state.value = LoginState.Error("Login failed: HTTP ${response.code()}")
                }
            } catch (e: Exception) {
                _state.value = LoginState.Error(e.message ?: "Unknown error occurred")
            }
        }
    }

    /**
     * 发送 FCM Token 到服务器
     */
    private fun submitFcmToken(baseUrl: String, jwtToken: String) {
        viewModelScope.launch {
            try {
                val fcmToken = preferenceManager?.getFcmToken()

                if (fcmToken != null) {
                    val authService = ApiClient.createAuthorizedClient(baseUrl, jwtToken)
                        .create(AuthService::class.java)

                    val request = FcmTokenRequest(
                        token = fcmToken,
                        deviceName = DeviceInfoUtils.getDeviceName()
                    )

                    val response = authService.submitFcmToken(request)

                    if (response.isSuccessful) {
                        Log.d(TAG, "FCM Token sent to server successfully")
                    } else {
                        Log.e(TAG, "Failed to send FCM Token: HTTP ${response.code()}")
                    }
                } else {
                    Log.w(TAG, "FCM Token is null, skipping submission")
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error sending FCM Token", e)
            }
        }
    }
}
