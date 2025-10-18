package com.fitonex.app.data.model

import kotlinx.serialization.Serializable

@Serializable
data class Checkin(
    val id: String,
    val user_id: String,
    val day: String,
    val created_at: String
)

@Serializable
data class CheckinStats(
    val current_streak_days: Int,
    val longest_streak_days: Int,
    val last_checkin_day: String? = null
)

@Serializable
data class CheckinTodayResponse(
    val checkin: Checkin,
    val inserted: Boolean
)
