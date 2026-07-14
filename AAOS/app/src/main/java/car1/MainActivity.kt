package car1

import android.graphics.Color
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.view.ViewGroup
import android.widget.*
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.google.gson.GsonBuilder
import com.google.gson.JsonObject
import kotlinx.coroutines.launch
import rust.interop.bridge.*

class MainActivity : AppCompatActivity() {

    private lateinit var statusTv: TextView
    private lateinit var dataContainer: LinearLayout
    private lateinit var prevBtn: Button
    private lateinit var nextBtn: Button
    private lateinit var loadingBar: ProgressBar
    private lateinit var rawJsonTv: TextView
    private lateinit var rawJsonContainer: LinearLayout

    private var currentPage = 1
    private var totalPages = 1
    private val prettyGson = GsonBuilder().setPrettyPrinting().create()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Root Horizontal Layout
        val rootLayout = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            setBackgroundColor(Color.BLACK)
            layoutParams = ViewGroup.LayoutParams(-1, -1)
            weightSum = 10f
        }

        // --- LEFT PANEL (30%) ---
        val leftPanel = createLeftPanel()
        rootLayout.addView(leftPanel)

        // --- RIGHT PANEL (70%) ---
        val rightPanel = createRightPanel()
        rootLayout.addView(rightPanel)

        setContentView(rootLayout)
        setupListeners()
        loadData()
    }

    private fun createLeftPanel() = LinearLayout(this).apply {
        orientation = LinearLayout.VERTICAL
        layoutParams = LinearLayout.LayoutParams(0, -1, 3f)
        setPadding(30, 40, 30, 40)
        gravity = Gravity.TOP or Gravity.CENTER_HORIZONTAL

        addView(TextView(context).apply {
            text = "BHILANI Interop SDK"; setTextColor(Color.parseColor("#BB86FC"))
            textSize = 30f; typeface = Typeface.DEFAULT_BOLD; gravity = Gravity.CENTER
        })

        statusTv = TextView(context).apply {
            text = "Page $currentPage"; setTextColor(Color.LTGRAY)
            textSize = 18f; setPadding(0, 20, 0, 10); gravity = Gravity.CENTER
        }
        addView(statusTv)

        loadingBar = ProgressBar(context, null, android.R.attr.progressBarStyleHorizontal).apply {
            layoutParams = LinearLayout.LayoutParams(-1, 15).apply { setMargins(0, 0, 0, 20) }
            visibility = View.INVISIBLE; isIndeterminate = true
        }
        addView(loadingBar)

        prevBtn = createNavButton("PREV")
        nextBtn = createNavButton("NEXT")
        addView(prevBtn)
        addView(View(context).apply { layoutParams = LinearLayout.LayoutParams(1, 20) })
        addView(nextBtn)
    }

    private fun createRightPanel() = ScrollView(this).apply {
        layoutParams = LinearLayout.LayoutParams(0, -1, 7f)
        setPadding(10, 40, 40, 40)
        isVerticalScrollBarEnabled = false

        val innerLayout = LinearLayout(context).apply {
            orientation = LinearLayout.VERTICAL
        }

        dataContainer = LinearLayout(context).apply { orientation = LinearLayout.VERTICAL }
        innerLayout.addView(dataContainer)

        // Raw JSON Debug Section
        rawJsonContainer = LinearLayout(context).apply {
            orientation = LinearLayout.VERTICAL
            visibility = View.GONE
            setPadding(20, 40, 20, 20)

            rawJsonTv = TextView(context).apply {
                setTextColor(Color.parseColor("#CCCCCC"))
                textSize = 12f
                typeface = Typeface.MONOSPACE
                setBackgroundColor(Color.parseColor("#121212"))
                setPadding(20, 20, 20, 20)
            }
            addView(rawJsonTv)
        }

        val toggleBtn = Button(context).apply {
            text = "TOGGLE RAW JSON"; textSize = 12f
            setOnClickListener {
                rawJsonContainer.visibility = if (rawJsonContainer.visibility == View.VISIBLE) View.GONE else View.VISIBLE
            }
        }
        innerLayout.addView(toggleBtn)
        innerLayout.addView(rawJsonContainer)

        addView(innerLayout)
    }

    private fun loadData() {
        updateLoadingState(true)
        lifecycleScope.launch {
            try {
                // Reusing the same logic file as standard Android!
                val result = fetchDataFromRust(currentPage)

                totalPages = result.pagination?.totalPages?.toInt() ?: 1
                updateUI(result)
                updateRawJson(result)

            } catch (e: Exception) {
                statusTv.text = "Sync Error"
            } finally {
                updateLoadingState(false)
            }
        }
    }

    private fun updateUI(result: FilterResponse) {
        dataContainer.removeAllViews()
        statusTv.text = "Page $currentPage of $totalPages"
        prevBtn.isEnabled = currentPage > 1
        nextBtn.isEnabled = currentPage < totalPages
        prevBtn.alpha = if (prevBtn.isEnabled) 1.0f else 0.3f
        nextBtn.alpha = if (nextBtn.isEnabled) 1.0f else 0.3f
        result.data.forEach { dataContainer.addView(createDataCard(it)) }
    }

    private fun updateRawJson(result: FilterResponse) {
        val orderedJson = JsonObject().apply {
            addProperty("message", result.message)
            add("pagination", prettyGson.toJsonTree(result.pagination))
            add("data", prettyGson.toJsonTree(result.data))
        }
        rawJsonTv.text = prettyGson.toJson(orderedJson)
    }

    private fun updateLoadingState(loading: Boolean) {
        loadingBar.visibility = if (loading) View.VISIBLE else View.INVISIBLE
        if (loading) statusTv.text = "Loading..."
    }

    private fun createNavButton(label: String) = Button(this).apply {
        text = label; textSize = 18f; typeface = Typeface.DEFAULT_BOLD
        layoutParams = LinearLayout.LayoutParams(-1, 110)
    }

    private fun createDataCard(item: Interoperability) = LinearLayout(this).apply {
        orientation = LinearLayout.VERTICAL
        setPadding(30, 25, 30, 25)
        background = GradientDrawable().apply {
            setColor(Color.parseColor("#1A1A1A"))
            cornerRadius = 12f; setStroke(2, Color.parseColor("#333333"))
        }
        layoutParams = LinearLayout.LayoutParams(-1, -2).apply { setMargins(0, 0, 0, 20) }

        addView(TextView(context).apply {
            text = item.title; setTextColor(Color.WHITE); textSize = 20f; typeface = Typeface.DEFAULT_BOLD
        })
        addView(TextView(context).apply {
            text = "Integration: ${item.integration}"; setTextColor(Color.parseColor("#03DAC6")); textSize = 16f
        })
    }

    private fun setupListeners() {
        prevBtn.setOnClickListener { if (currentPage > 1) { currentPage--; loadData() } }
        nextBtn.setOnClickListener { if (currentPage < totalPages) { currentPage++; loadData() } }
    }
}