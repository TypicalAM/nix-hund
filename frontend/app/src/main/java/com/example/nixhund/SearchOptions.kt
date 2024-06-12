package com.example.nixhund

import android.util.Log
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.nixhund.api.ApiClient
import com.example.nixhund.api.IndexInfo
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
            val channelNames = apiClient.getChannelList().channels
            val channelList = channelNames.map { name ->
                val indices = apiClient.getChannelIndices(name)
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