package com.fitonex.app.data.model

import kotlinx.serialization.Serializable

@Serializable
data class Gym(
    val id: String,
    val name: String,
    val lat: Double,
    val lng: Double,
    val address: String,
    val phone: String? = null,
    val website: String? = null,
    val created_at: String,
    val distance_m: Double? = null,
    val avg_rating: Double? = null,
    val machines_count: Int = 0,
    val price_from_cents: Int? = null
)

@Serializable
data class GymPrice(
    val gym_id: String,
    val plan_name: String,
    val price_cents: Int,
    val period: String
)

@Serializable
data class GymReview(
    val id: String,
    val gym_id: String,
    val user_id: String,
    val rating: Int,
    val comment: String? = null,
    val created_at: String,
    val user_name: String? = null
)

@Serializable
data class CreateReviewRequest(
    val rating: Int,
    val comment: String
)

@Serializable
data class GymReviewsResponse(
    val reviews: List<GymReview>,
    val next: String? = null
)
