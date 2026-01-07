package com.example.myapp.model

data class ApiResponse(
    val body: Body,
    val code: Int,
    val error: String,
    val message: String,
    val status: String
) {
    data class Body(
        val status: String
    )
}

