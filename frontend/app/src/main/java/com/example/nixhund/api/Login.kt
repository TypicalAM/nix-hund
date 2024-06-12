package com.example.nixhund.api

import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.addDefaultResponseValidation
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
data class LoginInfo(val username: String, val password: String)

@Serializable
data class LoginResponse(val token: String)

class LoginClient {
    private val baseUrl = "https://hund.piaseczny.dev"
    private val client = HttpClient(CIO) {
        install(ContentNegotiation) {
            addDefaultResponseValidation()
            json(Json {
                prettyPrint = true
                isLenient = true
                ignoreUnknownKeys = true
            })
        }
    }

    suspend fun register(info: LoginInfo): LoginResponse {
        val resp = client.post("$baseUrl/account/register") {
            contentType(ContentType.Application.Json)
            setBody(info)
        }

        if (resp.status.value > 400) {
            val msg: String = resp.body()
            throw Exception(msg)
        }

        return resp.body()
    }

    suspend fun login(info: LoginInfo): LoginResponse {
        val resp = client.post("$baseUrl/account/login") {
            contentType(ContentType.Application.Json)
            setBody(info)
        }

        if (resp.status.value > 400) {
            val msg: String = resp.body()
            throw Exception(msg)
        }

        return resp.body()
    }
}
