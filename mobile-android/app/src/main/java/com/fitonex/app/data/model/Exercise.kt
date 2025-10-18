package com.fitonex.app.data.model

import kotlinx.serialization.Serializable

@Serializable
data class Set(
    val id: String,
    val exercise_id: String,
    val set_index: Int,
    val reps: Int,
    val weight_kg: Double? = null,
    val rpe: Double? = null,
    val notes: String? = null
)

@Serializable
data class Exercise(
    val id: String,
    val user_id: String,
    val gym_id: String? = null,
    val machine_id: String? = null,
    val name: String,
    val created_at: String,
    val sets: List<Set> = emptyList(),
    val gym_name: String? = null,
    val machine_name: String? = null
)

@Serializable
data class CreateExerciseRequest(
    val day: String,
    val gym_id: String? = null,
    val machine_id: String? = null,
    val name: String,
    val sets: List<Set>
)
