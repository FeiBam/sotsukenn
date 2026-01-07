package com.example.myapp.network

import com.example.myapp.model.*
import retrofit2.Response
import retrofit2.http.*

/**
 * API Service - 所有 API 接口
 * 包含认证、摄像头等功能
 */
interface ApiService {

    // ========== 认证相关 ==========

    @GET("api/health")
    suspend fun health(): Response<ApiResponse>

    @POST("api/auth/login")
    suspend fun login(@Body request: LoginRequest): Response<LoginResponse>

    @POST("api/fcm/tokens")
    suspend fun submitFcmToken(@Body request: FcmTokenRequest): Response<FcmTokenResponse>

    // ========== 摄像头相关 ==========

    /**
     * 获取摄像头列表
     */
    @GET("api/cameras")
    suspend fun getCameras(): Response<CameraListResponse>

    /**
     * 获取摄像头快照 URL
     * @param cameraName 摄像头名称
     */
    @GET("api/camera/{camera_name}/snapshot")
    suspend fun getCameraSnapshot(
        @Path("camera_name") cameraName: String
    ): Response<CameraSnapshotResponse>

    /**
     * 获取摄像头流 URL
     * @param cameraName 摄像头名称
     * @param type 流类型 (mjpeg, hls, etc.)
     */
    @GET("api/camera/{camera_name}/stream")
    suspend fun getCameraStream(
        @Path("camera_name") cameraName: String,
        @Query("type") type: String = "mjpeg"
    ): Response<CameraStreamResponse>
}
