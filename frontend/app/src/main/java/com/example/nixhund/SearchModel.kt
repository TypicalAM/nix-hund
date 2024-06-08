package com.example.nixhund

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn

class SearchModel : ViewModel() {
    private val tests = listOf("hello", "everybody", "this", "is", "markiplier")
    private val _isSearching = MutableStateFlow(false)
    val isSearching = _isSearching.asStateFlow();

    private val _searchText = MutableStateFlow("")
    val searchText = _searchText.asStateFlow()

    private val _countriesList = MutableStateFlow(tests)
    val countriesList = searchText.combine(_countriesList) { text, countries ->
        if (text.isBlank()) {
            countries
        }
        countries.filter { country -> country.uppercase().contains(text.trim().uppercase()) }
    }.stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5000),
        initialValue = _countriesList.value
    )
}