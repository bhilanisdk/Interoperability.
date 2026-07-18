# Scala

**BhilaniSDK | Interoperability** by **Kantini, Chanchali**

*Get SDK Sample*

	https://github.com/bhilanisdk

*Run SDK Sample*

    scala-cli src/
    
*Basic Usage*

    //> using dep "com.lihaoyi::upickle:4.4.3"
    //> using javaOpt "--enable-native-access=ALL-UNNAMED"
    //> using javaProp "java.library.path=."
    
    package bhilani.interoperability.jvm
    
    import scala.util.Try
    import upickle.default.*
    
    @native def fetchInteroperability(url: String, paramsJson: String): String
    
    @main def runScala(): Unit =
      System.loadLibrary("interoperability_wrapper_robusta")
    
      val paramsMap = Map(
        "language"       -> null,
        "integration"    -> null,
        "crates"         -> null,
        "developmentkit" -> null,
        "page"           -> "1",
        "ids"            -> null
      )
      
      val paramsJson = write(paramsMap)
    
      println("--- Scala SDK ---")
    
      Try(fetchInteroperability("", paramsJson)).fold(
        e => println(s"Error: ${e.getMessage}"),
        res => println(res)
      )
      
*Dynamic Usage*

    //> using dep "com.lihaoyi::upickle:4.4.3"
    //> using javaOpt "--enable-native-access=ALL-UNNAMED"
    //> using javaProp "java.library.path=."
    
    package bhilani.interoperability.jvm
    
    import upickle.default.*
    import upickle.implicits.key
    import scala.util.Try
    
    case class Pagination(
      @key("total_pages") totalPages: Int
    ) derives ReadWriter
    
    case class SDKItem(
      title: String
    ) derives ReadWriter
    
    case class FetchResponse(
      data: List[SDKItem],
      pagination: Pagination
    ) derives ReadWriter
    
    class JVMSDKit:
      @native def fetchInteroperability(url: String, paramsJson: String): String
      
      System.loadLibrary("interoperability_wrapper_robusta")
    
      def fetchPage(url: String, page: Int): String =
        val params = write(Map("page" -> page.toString))
        fetchInteroperability(url, params)
    
    @main def runSDK(): Unit =
      val sdk = JVMSDKit()
      val url = ""
    
      println("--- Bhilani Interop SDK (Scala) ---")
    
      for pageNum <- 1 to 5 do
        val result = for
          response <- Try(sdk.fetchPage(url, pageNum))
          parsed   <- Try(read[FetchResponse](response))
        yield parsed
    
        result.fold(
          error => println(s"Page $pageNum: Failed (Error: ${error.getMessage})"),
          parsed => 
            val totalPages = parsed.pagination.totalPages
            if parsed.data.isEmpty || pageNum > totalPages then
              println(s"Page $pageNum: Success (No Data - Server has $totalPages pages)")
            else
              println(s"Page $pageNum: Success")
              parsed.data.foreach(item => println(s"  - Title: ${item.title}"))
        )

*Concurrent Usage*

    //> using dep "com.lihaoyi::upickle:4.4.3"
    //> using javaOpt "--enable-native-access=ALL-UNNAMED"
    //> using javaProp "java.library.path=."
    
    package bhilani.interoperability.jvm
    
    import upickle.default.*
    import upickle.implicits.key
    import scala.concurrent.{Future, Await}
    import scala.concurrent.duration.*
    import scala.concurrent.ExecutionContext.Implicits.global
    import scala.util.{Try, Success, Failure, Random}
    
    case class Pagination(
      @key("total_pages") totalPages: Int
    ) derives ReadWriter
    
    case class SDKItem(
      title: String
    ) derives ReadWriter
    
    case class FetchResponse(
      data: List[SDKItem],
      pagination: Pagination
    ) derives ReadWriter
    
    class JVMSDKit:
      @native private def fetchInteroperability(url: String, paramsJson: String): String
      System.loadLibrary("interoperability_wrapper_robusta")
    
      def fetchPages(url: String, pageRange: Range): Future[List[Try[String]]] =
        val tasks = pageRange.toList.map { page =>
          Future:
            Thread.sleep(Random.nextInt(201) + 50)
            val params = write(Map("page" -> page.toString))
            Try(Await.result(Future(fetchInteroperability(url, params)), 5.seconds))
        }
        Future.sequence(tasks)
    
    @main def runConcurrentSDK(): Unit =
      val sdk = JVMSDKit()
      println("--- Bhilani Interop SDK (Scala Concurrency) ---")
    
      val results = Await.result(sdk.fetchPages("", 1 to 5), 30.seconds)
    
      results.zipWithIndex.foreach { (result, index) =>
        val pageNum = index + 1
        result match
          case Success(res) =>
            Try(read[FetchResponse](res)) match
              case Success(parsed) =>
                val totalPages = parsed.pagination.totalPages
                if parsed.data.isEmpty || pageNum > totalPages then
                  println(s"Page $pageNum: Success (No Data - Server has $totalPages pages)")
                else
                  println(s"Page $pageNum: Success")
                  parsed.data.foreach(item => println(s"  - Title: ${item.title}"))
              case Failure(e) =>
                println(s"Page $pageNum: Success (JSON Parsing Failed: ${e.getMessage})")
          case Failure(error) =>
            println(s"Page $pageNum: Failed (${error.getMessage})")
      }

First time
<img width="934" height="436" alt="scala1" src="https://github.com/bhilanisdk/media/blob/main/scala1.png" />
Second time
<img width="953" height="434" alt="scala2" src="https://github.com/bhilanisdk/media/blob/main/scala2.png" />
Third time
<img width="922" height="440" alt="scala3" src="https://github.com/bhilanisdk/media/blob/main/scala3.png" />

**🙏 Mata Shabari 🙏**
