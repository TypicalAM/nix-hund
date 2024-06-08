package com.example.nixhund

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.core.IOException
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.emptyPreferences
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.runBlocking

val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "settings")

val USERNAME = stringPreferencesKey("username")
val LOGGED_IN = booleanPreferencesKey("logged_in")
val API_KEY = stringPreferencesKey("api_key")

fun getPreferenceString(context: Context, key: Preferences.Key<String>): Flow<String> {
    return context.dataStore.data.catch { exception ->
        if (exception is IOException) {
            emit(emptyPreferences())
        } else {
            throw exception
        }
    }.map { preferences ->
        preferences[key] ?: ""
    }
}

fun getPreferenceBool(context: Context, key: Preferences.Key<Boolean>): Flow<Boolean> {
    return context.dataStore.data.catch { exception ->
        if (exception is IOException) {
            emit(emptyPreferences())
        } else {
            throw exception
        }
    }.map { preferences ->
        preferences[key] ?: false
    }
}

fun getUsername(context: Context): String {
    return runBlocking { getPreferenceString(context, USERNAME).first() }
}

fun getApiKey(context: Context): String {
    return runBlocking { getPreferenceString(context, API_KEY).first() }
}

fun getLoggedIn(context: Context): Boolean {
    return runBlocking { getPreferenceBool(context, LOGGED_IN).first() }
}

suspend fun <T> setPref(context: Context, key: Preferences.Key<T>, value: T) {
    context.dataStore.edit { settings -> settings[key] = value }
}

fun setUsername(context: Context, value: String) {
    runBlocking { setPref(context, USERNAME, value) }
}

fun setApiKey(context: Context, value: String) {
    runBlocking { setPref(context, API_KEY, value) }
}

fun setLoggedIn(context: Context, value: Boolean) {
    runBlocking { setPref(context, LOGGED_IN, value) }
}