package com.fitonex.app.data.model

import kotlinx.serialization.Serializable

@Serializable
data class Machine(
    val id: String,
    val name: String,
    val body_part: String,
    val created_at: String
)

@Serializable
data class SearchMachinesResponse(
    val machines: List<Machine>
)
