package com.example.myapp.network

import android.os.Handler
import android.os.Looper
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.io.IOException
import java.util.concurrent.TimeUnit

object ApiClient {

    private const val BASE_URL = "http://localhost/" // Default, will be overridden

    var authHandler: AuthHandler? = null

    private val loggingInterceptor = HttpLoggingInterceptor().apply {
        level = HttpLoggingInterceptor.Level.BODY
    }

    private fun createUnauthorizedInterceptor(): Interceptor {
        return Interceptor { chain ->
            val response = chain.proceed(chain.request())
            if (response.code == 401) {
                // Close response to avoid resource leak
                response.close()

                // Trigger logout callback on main thread
                Handler(Looper.getMainLooper()).post {
                    authHandler?.onUnauthorized()
                }

                // Throw exception to prevent the response from being processed
                throw IOException("Unauthorized: Session expired or invalid token")
            }
            response
        }
    }

    private val okHttpClient = OkHttpClient.Builder()
        .addInterceptor(loggingInterceptor)
        .connectTimeout(30, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .writeTimeout(30, TimeUnit.SECONDS)
        .build()

    fun createClient(baseUrl: String): Retrofit {
        return Retrofit.Builder()
            .baseUrl(baseUrl)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }

    /**
     * 创建带 JWT Token 授权的 Retrofit 客户端
     * 用于需要身份验证的 API 请求
     */
    fun createAuthorizedClient(baseUrl: String, token: String): Retrofit {
        val authorizedClient = OkHttpClient.Builder()
            .addInterceptor(loggingInterceptor)
            .addInterceptor { chain ->
                val originalRequest = chain.request()
                val requestBuilder = originalRequest.newBuilder()
                    .header("Authorization", "Bearer $token")
                    .method(originalRequest.method, originalRequest.body)
                chain.proceed(requestBuilder.build())
            }
            .addInterceptor(createUnauthorizedInterceptor()) // Add 401 handler
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .build()

        return Retrofit.Builder()
            .baseUrl(baseUrl)
            .client(authorizedClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }
}
