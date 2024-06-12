package com.example.nixhund.api

import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.addDefaultResponseValidation
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType.Application.Json
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

class ApiClient(private val apiToken: String) {
    private val baseUrl = "https://hund.piaseczny.dev"
    private val client = HttpClient(CIO) {
        engine { requestTimeout = 0 }
        install(ContentNegotiation) {
            addDefaultResponseValidation()
            json(Json {
                prettyPrint = true
                isLenient = true
                ignoreUnknownKeys = true
            })
        }
    }

    suspend fun getChannelList(): ChannelList {
        val resp = client.get("$baseUrl/pkg/channel") {
            header("Authorization", "Bearer $apiToken")
        }

        if (resp.status.value > 400) {
            val msg: String = resp.body()
            throw Exception(msg)
        }

        return resp.body()
    }

    suspend fun getChannelIndices(channelId: String): List<IndexInfo> {
        val resp = client.get("$baseUrl/pkg/channel/index") {
            header("Authorization", "Bearer $apiToken")
            url { parameters.append("channel", channelId) }
        }

        if (resp.status.value > 400) {
            val msg: String = resp.body()
            throw Exception(msg)
        }

        return resp.body()
    }

    suspend fun generateIndex(channel: String): IndexGenerateResult {
        val resp = client.post("$baseUrl/pkg/channel/index/generate") {
            header("Authorization", "Bearer $apiToken")
            contentType(Json)
            setBody(IndexGenerateInput(channel))
        }

        if (resp.status.value > 400) {
            val msg: String = resp.body()
            throw Exception(msg)
        }

        return resp.body()
    }

    suspend fun indexQuery(id: String, query: String): List<PkgResult> {
        val resp = client.get("$baseUrl/pkg/index/$id/query") {
            header("Authorization", "Bearer $apiToken")
            url { parameters.append("query", query) }
        }

        if (resp.status.value > 400) {
            val msg: String = resp.body()
            throw Exception(msg)
        }

        return resp.body()
    }

    suspend fun getHistoryList(): List<HistoryEntry> {
        val resp = client.get("$baseUrl/account/history") {
            header("Authorization", "Bearer $apiToken")
        }

        if (resp.status.value > 400) {
            val msg: String = resp.body()
            throw Exception(msg)
        }

        return resp.body()
    }

    suspend fun deleteHistoryEntry(input: HistoryDeleteInput) {
        val resp = client.post("$baseUrl/account/history/delete") {
            header("Authorization", "Bearer $apiToken")
            setBody(input)
        }

        if (resp.status.value > 400) {
            val msg: String = resp.body()
            throw Exception(msg)
        }
    }
}