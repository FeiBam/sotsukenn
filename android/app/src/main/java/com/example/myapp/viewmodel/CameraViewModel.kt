package com.example.myapp.viewmodel

import android.util.Log
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.myapp.model.CameraSnapshotResponse
import com.example.myapp.network.ApiClient
import com.example.myapp.network.ApiService
import com.example.myapp.utils.PreferenceManager
import kotlinx.coroutines.launch

sealed class CameraState {
    object Idle : CameraState()
    object Loading : CameraState()
    data class CamerasLoaded(val cameras: List<String>) : CameraState()
    data class SnapshotLoaded(val snapshotUrl: String, val frigateToken: String) : CameraState()
    data class StreamLoaded(val streamUrl: String, val token: String) : CameraState()
    data class Error(val message: String) : CameraState()
}

class CameraViewModel : ViewModel() {

    private val _state = MutableLiveData<CameraState>(CameraState.Idle)
    val state: LiveData<CameraState> = _state

    private var preferenceManager: PreferenceManager? = null

    companion object {
        private const val TAG = "CameraViewModel"
    }

    fun setPreferenceManager(preferenceManager: PreferenceManager) {
        this.preferenceManager = preferenceManager
    }

    /**
     * 加载摄像头列表
     */
    fun loadCameras() {
        viewModelScope.launch {
            _state.value = CameraState.Loading

            try {
                val serverUrl = preferenceManager?.getServerUrl()
                val token = preferenceManager?.getToken()

                if (serverUrl != null && token != null) {
                    val apiService = ApiClient.createAuthorizedClient(serverUrl, token)
                        .create(ApiService::class.java)

                    val response = apiService.getCameras()

                    if (response.isSuccessful) {
                        val body = response.body()
                        if (body?.status == "success") {
                            _state.value = CameraState.CamerasLoaded(body.body.cameras)
                            Log.d(TAG, "Loaded ${body.body.cameras.size} cameras")
                        } else {
                            _state.value = CameraState.Error(body?.message ?: "Failed to load cameras")
                        }
                    } else {
                        _state.value = CameraState.Error("HTTP ${response.code()}")
                    }
                } else {
                    _state.value = CameraState.Error("Not logged in")
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error loading cameras", e)
                _state.value = CameraState.Error(e.message ?: "Unknown error")
            }
        }
    }

    /**
     * 获取摄像头快照 URL 和 frigate token
     */
    fun loadCameraSnapshot(cameraName: String) {
        viewModelScope.launch {
            try {
                val serverUrl = preferenceManager?.getServerUrl()
                val token = preferenceManager?.getToken()

                if (serverUrl != null && token != null) {
                    val apiService = ApiClient.createAuthorizedClient(serverUrl, token)
                        .create(ApiService::class.java)

                    val response = apiService.getCameraSnapshot(cameraName)

                    if (response.isSuccessful) {
                        val body = response.body()
                        if (body?.status == "success") {
                            _state.value = CameraState.SnapshotLoaded(
                                body.body.url,
                                body.body.frigateToken
                            )
                            Log.d(TAG, "Snapshot loaded: ${body.body.url}")
                        } else {
                            Log.e(TAG, "Failed to load snapshot: ${body?.message}")
                        }
                    } else {
                        Log.e(TAG, "Failed to load snapshot: HTTP ${response.code()}")
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error loading snapshot", e)
            }
        }
    }

    /**
     * 获取摄像头流 URL
     */
    fun loadCameraStream(cameraName: String) {
        viewModelScope.launch {
            _state.value = CameraState.Loading

            try {
                val serverUrl = preferenceManager?.getServerUrl()
                val token = preferenceManager?.getToken()

                if (serverUrl != null && token != null) {
                    val apiService = ApiClient.createAuthorizedClient(serverUrl, token)
                        .create(ApiService::class.java)

                    val response = apiService.getCameraStream(cameraName, "mjpeg")

                    if (response.isSuccessful) {
                        val body = response.body()
                        if (body?.status == "success") {
                            _state.value = CameraState.StreamLoaded(body.body.url, body.body.frigateToken)
                            Log.d(TAG, "Stream loaded: ${body.body.url}")
                        } else {
                            _state.value = CameraState.Error(body?.message ?: "Failed to load stream")
                        }
                    } else {
                        _state.value = CameraState.Error("HTTP ${response.code()}")
                    }
                } else {
                    _state.value = CameraState.Error("Not logged in")
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error loading stream", e)
                _state.value = CameraState.Error(e.message ?: "Unknown error")
            }
        }
    }

    /**
     * 获取摄像头快照 URL（旧方法，已弃用）
     */
    fun getSnapshotUrl(cameraName: String): String? {
        val serverUrl = preferenceManager?.getServerUrl()
        return if (serverUrl != null) {
            "$serverUrl/api/camera/$cameraName/snapshot"
        } else {
            null
        }
    }

    /**
     * 获取带认证的快照请求头
     */
    fun getSnapshotHeaders(): Map<String, String> {
        val token = preferenceManager?.getToken()
        return if (token != null) {
            mapOf("Authorization" to "Bearer $token")
        } else {
            emptyMap()
        }
    }
}
