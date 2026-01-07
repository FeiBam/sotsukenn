package com.example.myapp.network

import com.example.myapp.model.ApiResponse
import com.example.myapp.model.FcmTokenRequest
import com.example.myapp.model.FcmTokenResponse
import com.example.myapp.model.LoginRequest
import com.example.myapp.model.LoginResponse
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST

interface AuthService {

    @GET("api/health")
    suspend fun health(): Response<ApiResponse>

    @POST("api/auth/login")
    suspend fun login(@Body request: LoginRequest): Response<LoginResponse>

    @POST("api/fcm/tokens")
    suspend fun submitFcmToken(@Body request: FcmTokenRequest): Response<FcmTokenResponse>
}
