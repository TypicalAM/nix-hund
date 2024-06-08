package com.example.nixhund.screens

import androidx.compose.material3.Button
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.navigation.NavHostController

@Composable
fun History(navHostController: NavHostController) {
    Button(onClick = { navHostController.navigate("search")}) {
        Text(text = "history")
    }
}