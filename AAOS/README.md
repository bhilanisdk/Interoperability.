# Android AAOS

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
<img width="1080" height="600" alt="aaos1" src="https://github.com/bhilanisdk/media/blob/main/aaos1.png" />

Screenshot (Page 4)
<img width="1080" height="600" alt="aaos1" src="https://github.com/bhilanisdk/media/blob/main/aaos2.png" />
<img width="1080" height="600" alt="aaos2" src="https://github.com/user-attachments/assets/edb50c96-0e0c-46ce-a609-d407f0c6c0be" />

**🙏 Mata Shabari 🙏**
