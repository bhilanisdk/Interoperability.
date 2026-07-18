# Java

**BhilaniSDK | Interoperability** by **Kantini, Chanchali**

*Get SDK Sample*

	https://github.com/bhilanisdk

*Run SDK Sample*

    java --enable-preview --enable-native-access=ALL-UNNAMED -Djava.library.path=. src/bhilani/interoperability/jvm/JVMSDKit.java

*Basic Usage*

    package bhilani.interoperability.jvm;
    
    public class JVMSDKit {
    
        private native String fetchInteroperability(String url, String paramsJson);
    
        static {
            System.loadLibrary("interoperability_wrapper_robusta");
        }
    
        void main() {
            String params = """
                {
                  "language": null,
                  "integration": null,
                  "crates": null,
                  "developmentkit": null,
                  "page": "1",
                  "ids": null
                }
                """;
                
            System.out.println("Java SDK");
            
            System.out.println(fetchInteroperability("", params));
        }
    }

*Dynamic Usage*
  
    package bhilani.interoperability.jvm;
    
    import java.util.regex.Matcher;
    import java.util.regex.Pattern;
    
    public class JVMSDKit {
    
        public native String fetchInteroperability(String url, String paramsJson);
    
        static {
            System.loadLibrary("interoperability_wrapper_robusta");
        }
    
        public String fetchPage(String url, int page) {
            String params = String.format("{\"page\": \"%d\"}", page);
            return fetchInteroperability(url, params);
        }
    
        public static void main(String[] args) {
            JVMSDKit sdk = new JVMSDKit();
            String url = "";
    
            System.out.println("--- Bhilani Interop SDK ---");
    
            for (int pageNum = 1; pageNum <= 5; pageNum++) {
                try {
                    String response = sdk.fetchPage(url, pageNum);
                    
                    int totalPages = parseTotalPages(response);
                    var titles = parseTitles(response);
    
                    if (titles.isEmpty() || pageNum > totalPages) {
                        System.out.printf("Page %d: Success (No Data - Server has %d pages)%n", pageNum, totalPages);
                    } else {
                        System.out.printf("Page %d: Success%n", pageNum);
                        for (String title : titles) {
                            System.out.println("  - Title: " + title);
                        }
                    }
                } catch (Exception e) {
                    System.out.printf("Page %d: Failed (Error: %s)%n", pageNum, e.getMessage());
                }
            }
        }
    
        private static int parseTotalPages(String json) {
            Pattern pattern = Pattern.compile("\"total_pages\"\\s*:\\s*(\\d+)");
            Matcher matcher = pattern.matcher(json);
            return matcher.find() ? Integer.parseInt(matcher.group(1)) : 0;
        }
    
        private static java.util.List<String> parseTitles(String json) {
            java.util.List<String> titles = new java.util.ArrayList<>();
            Pattern pattern = Pattern.compile("\"title\"\\s*:\\s*\"([^\"]+)\"");
            Matcher matcher = pattern.matcher(json);
            while (matcher.find()) {
                titles.add(matcher.group(1));
            }
            return titles;
        }
    }

*Concurrent Usage*

    package bhilani.interoperability.jvm;
    
    import java.util.ArrayList;
    import java.util.List;
    import java.util.Random;
    import java.util.concurrent.*;
    import java.util.regex.Matcher;
    import java.util.regex.Pattern;
    import java.util.stream.Collectors;
    import java.util.stream.IntStream;
    
    public class JVMSDKit {
    
        private native String fetchInteroperability(String url, String paramsJson);
    
        static {
            System.loadLibrary("interoperability_wrapper_robusta");
        }
    
        public List<CompletableFuture<String>> fetchPages(String url, int start, int end) {
            ExecutorService executor = Executors.newCachedThreadPool();
            Random random = new Random();
    
            return IntStream.rangeClosed(start, end)
                .mapToObj(page -> CompletableFuture.supplyAsync(() -> {
                    try {
                        Thread.sleep(random.nextInt(201) + 50);
    
                        return CompletableFuture.supplyAsync(() -> 
                            fetchInteroperability(url, String.format("{\"page\": \"%d\"}", page))
                        ).get(5, TimeUnit.SECONDS);
    
                    } catch (Exception e) {
                        throw new CompletionException(e);
                    }
                }, executor))
                .collect(Collectors.toList());
        }
    
        public static void main(String[] args) {
            JVMSDKit sdk = new JVMSDKit();
            String url = "";
    
            System.out.println("--- Bhilani Interop SDK (Java Concurrency) ---");
    
            List<CompletableFuture<String>> futures = sdk.fetchPages(url, 1, 5);
    
            for (int i = 0; i < futures.size(); i++) {
                int pageNum = i + 1;
                try {
                    String response = futures.get(i).join();
                    
                    int totalPages = parseTotalPages(response);
                    List<String> titles = parseTitles(response);
    
                    if (titles.isEmpty() || pageNum > totalPages) {
                        System.out.printf("Page %d: Success (No Data - Server has %d pages)%n", pageNum, totalPages);
                    } else {
                        System.out.printf("Page %d: Success%n", pageNum);
                        titles.forEach(t -> System.out.println("  - Title: " + t));
                    }
                } catch (Exception e) {
                    System.out.printf("Page %d: Failed (%s)%n", pageNum, e.getCause().getMessage());
                }
            }
        }
    
        private static int parseTotalPages(String json) {
            Matcher m = Pattern.compile("\"total_pages\"\\s*:\\s*(\\d+)").matcher(json);
            return m.find() ? Integer.parseInt(m.group(1)) : 0;
        }
    
        private static List<String> parseTitles(String json) {
            List<String> titles = new ArrayList<>();
            Matcher m = Pattern.compile("\"title\"\\s*:\\s*\"([^\"]+)\"").matcher(json);
            while (m.find()) titles.add(m.group(1));
            return titles;
        }
    }

Fist time
<img width="967" height="447" alt="java1" src="https://github.com/bhilanisdk/media/blob/main/java1.png" />

Second time
<img width="942" height="438" alt="java2" src="https://github.com/bhilanisdk/media/blob/main/java2.png" />

Third time
<img width="868" height="440" alt="java3" src="https://github.com/bhilanisdk/media/blob/main/java3.png" />

**🙏 Mata Shabari 🙏**
