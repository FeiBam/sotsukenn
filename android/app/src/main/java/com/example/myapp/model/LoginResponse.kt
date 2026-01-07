package com.example.myapp.model

data class LoginResponse(
    val body: Body,
    val code: Int,
    val error: String,
    val message: String,
    val status: String
) {
    data class Body(
        val token: String,
        val user: UserInfo? = null
    )
}

data class UserInfo(
    val id: String? = null,
    val username: String? = null,
    val email: String? = null
)
