# Android

**BhilaniSDK | Interoperability** by **Kantini, Chanchali**

*Get SDK Sample*

	https://github.com/bhilanisdk

*Run SDK Sample*

    Android Studio

*Usage*

    import kotlinx.coroutines.Dispatchers
    import kotlinx.coroutines.withContext
    import rust.interop.bridge.FilterParams
    import rust.interop.bridge.FilterResponse
    import rust.interop.bridge.fetchInteroperability
    
    val params = FilterParams(
        integration = null,
        developmentkit = null,
        language = null,
        crates = null,
        page = "1",
        ids = null
    )
    
    suspend fun fetchDataFromRust(pageNumber: Int): FilterResponse {
        params.page = pageNumber.toString();
        return withContext(Dispatchers.IO) {
            fetchInteroperability(params)
        }
    }

Screenshot (Page 1)
<img width="1080" height="2340" alt="android1" src="https://github.com/bhilanisdk/media/blob/main/android1.jpg" />

Screenshot (Page 4)
<img width="1080" height="2340" alt="android2" src="https://github.com/bhilanisdk/media/blob/main/android2.jpg" />

**🙏 Mata Shabari 🙏**
