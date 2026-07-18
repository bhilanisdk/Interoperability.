# Groovy

**BhilaniSDK | Interoperability** by **Kantini, Chanchali**

*Get SDK Sample*

	https://github.com/bhilanisdk

*Run SDK Sample*

    Command 1: Compile the Java bridge

        javac --enable-preview --release 26 src/bhilani/interoperability/jvm/JVMSDKit.java
  
    Command 2: Set the JVM flags

        set JAVA_OPTS=--add-opens java.base/java.lang=ALL-UNNAMED --add-opens java.base/java.util=ALL-UNNAMED

    Command 3: Run the Groovy script using the compiled Java class

        groovy -cp src src/bhilani/interoperability/jvm/RunSDK.groovy


*Basic Usage*

    package bhilani.interoperability.jvm
    
    import groovy.json.*
    
    def sdk = new JVMSDKit()
    
    def params = [ 
        language: null,
        integration: null,
        crates: null,
        developmentkit: null,
        page: "1",
        ids: null 
    ]
    
    try {
        println "Groovy SDK"
        
        def jsonInput = JsonOutput.toJson(params)
        def response = sdk.fetchInteroperability("", jsonInput)
        
        println JsonOutput.prettyPrint(response)
    } catch (e) {
        println "Error: $e.message"
    }
    
*Dynamic Usage*

      import bhilani.interoperability.jvm.JVMSDKit
      import groovy.json.JsonOutput
      import groovy.json.JsonSlurper
      
      def sdk = new JVMSDKit()
      def slurper = new JsonSlurper()
      def url = ""
      
      println "--- Bhilani Interop SDK ---"
      
      (1..5).each { pageNum ->
          try {
      
              def params = [page: "${pageNum}"]
              def jsonInput = JsonOutput.toJson(params)
              
              def response = sdk.fetchInteroperability(url, jsonInput)
              
              def parsed = slurper.parseText(response)
              
              def totalPages = parsed.pagination.total_pages
              def dataItems = parsed.data
      
              if (!dataItems || pageNum > totalPages) {
                  println "Page $pageNum: Success (No Data - Server has $totalPages pages)"
              } else {
                  println "Page $pageNum: Success"
                  dataItems.each { item ->
                      println "  - Title: ${item.title}"
                  }
              }
          } catch (Exception e) {
              println "Page $pageNum: Failed (Error: ${e.message})"
          }
      }

*Concurrent Usage*

      @Grab('org.codehaus.gpars:gpars:1.2.1')
      import bhilani.interoperability.jvm.JVMSDKit
      import groovy.json.*
      import groovyx.gpars.GParsPool
      
      def parseResponse(String response) {
          return new JsonSlurper().parseText(response)
      }
      
      def sdk = new JVMSDKit()
      def rand = new Random()
      
      println "--- Bhilani Interop SDK (GPars Parallel) ---"
      
      GParsPool.withPool {
          def results = (1..5).collectParallel { pageNum ->
              try {
                  sleep(rand.nextInt(201) + 50)
                  
                  def params = JsonOutput.toJson([page: "${pageNum}"])
                  def response = sdk.fetchInteroperability("", params)
                  def parsed = parseResponse(response)
                  
                  return [
                      pageNum: pageNum, 
                      success: true, 
                      data: parsed.data, 
                      total: parsed.pagination.total_pages
                  ]
              } catch (e) {
                  return [pageNum: pageNum, success: false, error: e.message]
              }
          }
      
          results.each { res ->
              if (res.success) {
                  if (!res.data || res.pageNum > res.total) {
                      println "Page ${res.pageNum}: Success (No Data - Server has ${res.total} pages)"
                  } else {
                      println "Page ${res.pageNum}: Success"
                      res.data.each { println "  - Title: ${it.title}" }
                  }
              } else {
                  println "Page ${res.pageNum}: Failed (${res.error})"
              }
          }
      }

First time
<img width="1053" height="434" alt="groovy1" src="https://github.com/bhilanisdk/media/blob/main/groovy1.png" />

Second time
<img width="1050" height="432" alt="groovy2" src="https://github.com/bhilanisdk/media/blob/main/groovy2.png" />

Third time
<img width="1046" height="445" alt="groovy3" src="https://github.com/bhilanisdk/media/blob/main/groovy3.png" />

**🙏 Mata Shabari 🙏**
