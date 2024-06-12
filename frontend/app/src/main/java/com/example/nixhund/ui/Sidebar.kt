package com.example.nixhund.ui

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavHostController
import com.example.nixhund.R
import com.example.nixhund.getUsername
import com.example.nixhund.setApiKey
import com.example.nixhund.setLoggedIn
import com.example.nixhund.setPref
import com.example.nixhund.setUsername

@Composable
fun Sidebar(navHostController: NavHostController) {
    val scope = rememberCoroutineScope()
    val context = LocalContext.current
    val username = getUsername(context)

    Column(
        modifier = Modifier
            .fillMaxHeight()
            .width(200.dp)
            .background(Color.LightGray)
            .padding(16.dp),
        verticalArrangement = Arrangement.SpaceBetween,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Image(
                painter = painterResource(id = R.drawable.ic_launcher_foreground),
                contentDescription = null,
                modifier = Modifier
                    .size(80.dp)
                    .background(Color.White, CircleShape)
                    .padding(8.dp)
            )
            Text("nix hund", fontWeight = FontWeight.Bold, fontSize = 20.sp)
            Text("v1.0", fontSize = 14.sp)
            Spacer(modifier = Modifier.height(16.dp))
            SidebarButton("Home", onClick = { navHostController.navigate("search") })
            SidebarButton("History", onClick = { navHostController.navigate("history") })
            SidebarButton("Channel", onClick = { navHostController.navigate("channel") })
            SidebarButton("Index", onClick = { navHostController.navigate("index") })
            SidebarButton("Logout", onClick = {
                scope.run {
                    setApiKey(context, "")
                    setUsername(context, "")
                    setLoggedIn(context, false)
                    navHostController.navigate("welcome")
                }
            }, color = Color.Red)
        }
        Text("Logged in as $username", color = Color.Green, fontSize = 14.sp)
    }
}

@Composable
fun SidebarButton(text: String, color: Color = Color.Black, onClick: () -> Unit) {
    Button(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        colors = ButtonDefaults.buttonColors(),
        shape = RoundedCornerShape(8.dp)
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                painter = painterResource(id = R.drawable.ic_launcher_foreground),
                contentDescription = null,
                tint = color,
                modifier = Modifier.size(20.dp)
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(text, color = color)
        }
    }
}