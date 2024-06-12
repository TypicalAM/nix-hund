package com.example.nixhund

import android.util.Log
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.nixhund.api.ApiClient
import com.example.nixhund.api.IndexInfo
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

data class ChannelInfo(val name: String, val indices: List<IndexInfo>)

class SearchViewModel : ViewModel() {
    var channels by mutableStateOf<List<ChannelInfo>>(emptyList())
        private set

    var currentChannel by mutableStateOf<ChannelInfo?>(null)
    var currentIndex by mutableStateOf<IndexInfo?>(null)
    var populated by mutableStateOf(false)

    fun populateData(apiClient: ApiClient) {
        viewModelScope.launch {
            var channelNames: List<String> = listOf()
            try {
                channelNames = apiClient.getChannelList().channels
            } catch (e: Exception) {
                Log.d("search_model", "Exception in popuplate: $e")
                cancel()
            }
            val channelList = channelNames.map { name ->
                var indices: List<IndexInfo> = listOf()
                try {
                    indices = apiClient.getChannelIndices(name)
                } catch (e: Exception) {
                    Log.d("search_model", "Exception in popuplate: $e")
                    cancel()
                }
                Log.d("search_model", "$name has ${indices.size} indices")
                ChannelInfo(name, indices)
            }

            Log.d("search_model", "Populated channels with ${channels.size} entries")
            channels = channelList
            populated = true

            var found = false
            for (channel in channelList) if (channel.indices.isNotEmpty()) {
                currentChannel = channel
                currentIndex = channel.indices[0]
                found = true
                break
            }

            if (!found && channelList.isNotEmpty()) currentChannel = channelList[0]
        }
    }
}