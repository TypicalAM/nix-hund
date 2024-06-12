package com.example.nixhund.ui

import androidx.compose.foundation.layout.Row
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material3.Icon
import androidx.compose.material3.NavigationDrawerItem
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.navigation.NavHostController

@Composable
fun NavigationList(navHostController: NavHostController) {
    NavigationDrawerItem(
        label = {
            Row {
                Icon(imageVector = Icons.Filled.Home, contentDescription = "Home page")
                Text(text = "Search")
            }
        },
        selected = true,
        onClick = { navHostController.navigate("search") },
    )

    NavigationDrawerItem(
        label = {
            Row {
                Icon(imageVector = Icons.Filled.Home, contentDescription = "Home page")
                Text(text = "History")
            }
        },
        selected = false,
        onClick = { navHostController.navigate("history") },
    )
}