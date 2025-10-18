package com.fitonex.app.data.model

import kotlinx.serialization.Serializable

@Serializable
data class PaginatedResponse<T>(
    val items: List<T>,
    val next_cursor: String? = null,
    val has_more: Boolean = false
)
