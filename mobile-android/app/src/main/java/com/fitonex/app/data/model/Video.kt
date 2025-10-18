package com.fitonex.app.data.model

import kotlinx.serialization.Serializable

@Serializable
data class InstructionVideo(
    val id: String,
    val machine_id: String,
    val uploader_id: String,
    val title: String,
    val description: String? = null,
    val video_key: String,
    val thumb_key: String? = null,
    val duration_sec: Int? = null,
    val created_at: String,
    val like_count: Int = 0,
    val is_liked: Boolean = false,
    val uploader_name: String? = null,
    val machine_name: String? = null
)

@Serializable
data class UploadURLRequest(
    val machine_id: String,
    val title: String,
    val description: String,
    val content_type: String,
    val bytes: Long
)

@Serializable
data class UploadURLResponse(
    val upload_url: String,
    val video_key: String,
    val thumb_upload_url: String? = null,
    val thumb_key: String? = null
)

@Serializable
data class FinalizeVideoRequest(
    val machine_id: String,
    val title: String,
    val description: String,
    val video_key: String,
    val thumb_key: String? = null,
    val duration_sec: Int
)
