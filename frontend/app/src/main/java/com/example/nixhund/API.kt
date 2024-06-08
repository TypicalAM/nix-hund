package com.example.nixhund

import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.addDefaultResponseValidation
import io.ktor.client.request.*
import io.ktor.http.*
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.PrimitiveSerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
import kotlinx.serialization.json.Json
import java.util.Date

object DateSerializer : KSerializer<Date> {
    override val descriptor = PrimitiveSerialDescriptor("Date", PrimitiveKind.LONG)
    override fun serialize(encoder: Encoder, value: Date) = encoder.encodeLong(value.time)
    override fun deserialize(decoder: Decoder): Date = Date(decoder.decodeLong())
}

@Serializable
data class PkgResult(
    @SerialName("pkg_name") val pkgName: String,
    @SerialName("out_name") val outName: String,
    @SerialName("out_hash") val outHash: String,
    val path: String,
    val version: String
)

@Serializable
data class HistoryEntry(
    @SerialName("index_id") val indexID: String,
    @Serializable(DateSerializer::class) val date: Date,
    val pkg: PkgResult
)

@Serializable
data class IndexInfo(
    val id: String,
    @Serializable(DateSerializer::class) val date: Date,
    @SerialName("total_file_count") val totalFileCount: Int,
)

@Serializable
data class ChannelList(val channels: List<String>)

@Serializable
data class IndexGenerateInput(val channel: String)

@Serializable
data class HistoryDeleteInput(val index: String)

@Serializable
data class IndexGenerateResult(
    val id: String,
    val time: String,
    @SerialName("total_package_count") val totalPackageCount: Int,
    @SerialName("total_file_count") val totalFileCount: Int,
)

@Serializable
data class LoginInfo(
    val username: String, val password: String
)

@Serializable
data class LoginResponse(val token: String)

class LoginClient {
    private val client = HttpClient(CIO) {
        install(ContentNegotiation) {
            json(Json {
                prettyPrint = true
                isLenient = true
                ignoreUnknownKeys = true
            })
            addDefaultResponseValidation()
        }
    }
    private val baseUrl = "https://hund.piaseczny.dev"

    suspend fun register(info: LoginInfo): LoginResponse {
        return client.post("$baseUrl/account/register") {
            contentType(ContentType.Application.Json)
            setBody(info)
        }.body()
    }

    suspend fun login(info: LoginInfo): LoginResponse {
        return client.post("$baseUrl/account/login") {
            contentType(ContentType.Application.Json)
            setBody(info)
        }.body()
    }
}

class ApiClient(private val apiToken: String) {
    private val client = HttpClient(CIO) { install(ContentNegotiation) { json() } }
    private val baseUrl = "https://hund.piaseczny.dev"

    suspend fun getChannelList(): ChannelList {
        return client.get("$baseUrl/channel") {
            header("Authorization", "Bearer $apiToken")
        }.body()
    }

    suspend fun getChannelIndices(channelId: String): IndexInfo {
        return client.get("$baseUrl/index") {
            header("Authorization", "Bearer $apiToken")
            url { parameters.append("channel", channelId) }
        }.body()
    }

    suspend fun generateIndex(channel: String): IndexGenerateResult {
        return client.post("$baseUrl/index/generate") {
            header("Authorization", "Bearer $apiToken")
            contentType(ContentType.Application.Json)
            setBody(IndexGenerateInput(channel))
        }.body()
    }

    suspend fun indexQuery(id: String, query: String): List<PkgResult> {
        return client.get("$baseUrl/index/$id/query") {
            header("Authorization", "Bearer $apiToken")
            url { parameters.append("query", query) }
        }.body()
    }

    suspend fun getHistoryList(): List<HistoryEntry> {
        return client.get("$baseUrl/account/history") {
            header("Authorization", "Bearer $apiToken")
        }.body()
    }

    suspend fun deleteHistoryEntry(input: HistoryDeleteInput) {
        client.post("$baseUrl/account/history/delete") {
            header("Authorization", "Bearer $apiToken")
            setBody(input)
        }
    }

    suspend fun deleteUser() {
        client.post("$baseUrl/account/delete") { header("Authorization", "Bearer $apiToken") }
    }
}