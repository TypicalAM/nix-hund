package com.example.nixhund.api

import android.annotation.SuppressLint
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.addDefaultResponseValidation
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.PrimitiveSerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
import kotlinx.serialization.json.Json
import java.text.SimpleDateFormat
import java.util.Date

object DateSerializer : KSerializer<Date> {
    @SuppressLint("SimpleDateFormat")
    private val format = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSSSSSSSSZ")
    override val descriptor = PrimitiveSerialDescriptor("Date", PrimitiveKind.LONG)
    override fun serialize(encoder: Encoder, value: Date) =
        encoder.encodeString(format.format(value))

    override fun deserialize(decoder: Decoder): Date = format.parse(decoder.decodeString())!!
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
            contentType(ContentType.Application.Json)
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

    suspend fun deleteUser() {
        val resp =
            client.post("$baseUrl/account/delete") { header("Authorization", "Bearer $apiToken") }

        if (resp.status.value > 400) {
            val msg: String = resp.body()
            throw Exception(msg)
        }

        return resp.body()
    }
}