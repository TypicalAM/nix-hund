package com.example.nixhund.api

import android.annotation.SuppressLint
import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.PrimitiveSerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
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
