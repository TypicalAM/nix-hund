package com.example.nixhund.screens

import androidx.compose.material3.Button
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext
import androidx.navigation.NavHostController
import com.example.nixhund.setLoggedIn

@Composable
fun Settings(navHostController: NavHostController) {
    val context = LocalContext.current
    Button(onClick = { setLoggedIn(context, true) }) {
        Text(text = "Change to logged in")
    }
}