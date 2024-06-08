package com.example.nixhund.ui

import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.vectorResource
import androidx.compose.ui.tooling.preview.Preview
import com.example.nixhund.R
import com.example.nixhund.ui.theme.Pink80
import com.example.nixhund.ui.theme.Purple80

@Composable
fun ListEntryWithCopy(
    modifier: Modifier = Modifier,
    imageVector: ImageVector,
    text: String,
    onClick: (isEnabled: Boolean) -> Unit = {},
    enable: Boolean = true,
    backgroundColor: Color,
    fontColor: Color,
) {
    Text(text = text)
}

@Preview
@Composable
fun ListEntryWithCopyPreview() {/*val clipboardManager: ClipboardManager = LocalClipboardManager.current
var text by remember { mutableStateOf("")}

Column(modifier = Modifier.fillMaxSize()) {

    TextField(value = text, onValueChange = {text = it})
    Button(onClick = {
        clipboardManager.setText(AnnotatedString((text)))
    }) {
        Text("Copy")
    }

    Button(onClick = {
      clipboardManager.getText()?.text?.let {
          text = it
      }
    }) {
        Text("Get")
    }
}
*/
    ListEntryWithCopy(
        imageVector = ImageVector.vectorResource(R.drawable.baseline_content_copy_24),
        text = "perl532Packages.MooX...",
        enable = false,
        backgroundColor = Purple80,
        fontColor = Pink80,
    )
}