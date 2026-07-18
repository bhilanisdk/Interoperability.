# Csharp

**BhilaniSDK | Interoperability** by **Kantini, Chanchali**

*Get SDK Sample*

	https://github.com/bhilanisdk

*Run SDK Sample*

    dotnet run

*Basic Usage*

    using System.Text.Json;
    using SDK = InteroperabilityWrapperRnet.InteroperabilityWrapperRnet;
    
    string paramsJson = """
        {
          "language": null,
          "integration": null,
          "crates": null,
          "developmentkit": null,
          "page": "1",
          "ids": null
        }
        """;
    
    Console.WriteLine(".NET SDK");
    
    try 
    {
        string response = SDK.FetchForDotnet(paramsJson);
        Console.WriteLine(response);
    }
    catch (Exception e) 
    {
        Console.WriteLine($"Error: {e.Message}");
    }

*Dynamic Usage*

    using System.Text.Json;
    using System.Text.Json.Serialization;
    using SDK = InteroperabilityWrapperRnet.InteroperabilityWrapperRnet;
    
    public record Pagination(
        [property: JsonPropertyName("total_pages")] int TotalPages
    );
    
    public record SDKItem(
        [property: JsonPropertyName("title")] string Title
    );
    
    public record FetchResponse(
        [property: JsonPropertyName("data")] List<SDKItem> Data,
        [property: JsonPropertyName("pagination")] Pagination Pagination
    );
    
    public class DotnetSDKit
    {
        public string FetchPage(int page)
        {
            string paramsJson = $"{{\"page\": \"{page}\"}}";
            return SDK.FetchForDotnet(paramsJson);
        }
    }
    
    class Program
    {
        static void Main()
        {
            var sdk = new DotnetSDKit();
            var options = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
    
            Console.WriteLine("--- Bhilani Interop SDK ---");
    
            for (int pageNum = 1; pageNum <= 5; pageNum++)
            {
                try
                {
                    string response = sdk.FetchPage(pageNum);
                    var parsed = JsonSerializer.Deserialize<FetchResponse>(response, options);
    
                    int totalPages = parsed?.Pagination.TotalPages ?? 0;
    
                    if (parsed?.Data == null || parsed.Data.Count == 0 || pageNum > totalPages)
                    {
                        Console.WriteLine($"Page {pageNum}: Success (No Data - Server has {totalPages} pages)");
                    }
                    else
                    {
                        Console.WriteLine($"Page {pageNum}: Success");
                        foreach (var item in parsed.Data)
                        {
                            Console.WriteLine($"  - Title: {item.Title}");
                        }
                    }
                }
                catch (Exception e)
                {
                    Console.WriteLine($"Page {pageNum}: Failed (Error: {e.Message})");
                }
            }
        }
    }

*Concurrent Usage*

        using System.Text.Json;
        using System.Text.Json.Serialization;
        using SDK = InteroperabilityWrapperRnet.InteroperabilityWrapperRnet;
        
        namespace Bhilani.Interoperability;
        
        public record SDKItem(
            [property: JsonPropertyName("title")] string Title
        );
        
        public record Pagination(
            [property: JsonPropertyName("total_pages")] int TotalPages
        );
        
        public record FetchResponse(
            [property: JsonPropertyName("data")] List<SDKItem> Data,
            [property: JsonPropertyName("pagination")] Pagination Pagination
        );
        
        public record Result<T>(T? Value, Exception? Error)
        {
            public bool IsSuccess => Error == null;
            public static Result<T> Success(T value) => new(value, null);
            public static Result<T> Failure(Exception e) => new(default, e);
        }
        
        public class DotnetSDKit
        {
            public async Task<List<Result<string>>> FetchPagesAsync(IEnumerable<int> pageRange)
            {
                var tasks = pageRange.Select(async page =>
                {
                    await Task.Delay(Random.Shared.Next(50, 251));
        
                    try
                    {
                        using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(5));
                        var paramsJson = $"{{\"page\": \"{page}\"}}";
        
                        var response = await Task.Run(() => SDK.FetchForDotnet(paramsJson), cts.Token);
                        
                        return Result<string>.Success(response);
                    }
                    catch (Exception ex)
                    {
                        return Result<string>.Failure(ex);
                    }
                });
        
                return (await Task.WhenAll(tasks)).ToList();
            }
        }
        
        class Program
        {
            static async Task Main()
            {
                var sdk = new DotnetSDKit();
                var jsonOptions = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
        
                Console.WriteLine("--- Bhilani Interop SDK (.NET Concurrency) ---\n");
        
                var results = await sdk.FetchPagesAsync(Enumerable.Range(1, 5));
        
                for (int i = 0; i < results.Count; i++)
                {
                    int pageNum = i + 1;
                    var result = results[i];
        
                    if (result.IsSuccess && result.Value != null)
                    {
                        try
                        {
                            var parsed = JsonSerializer.Deserialize<FetchResponse>(result.Value, jsonOptions);
                            int totalPages = parsed?.Pagination.TotalPages ?? 0;
        
                            if (parsed?.Data == null || parsed.Data.Count == 0 || pageNum > totalPages)
                            {
                                Console.WriteLine($"Page {pageNum}: Success (No Data - Server has {totalPages} pages)");
                            }
                            else
                            {
                                Console.WriteLine($"Page {pageNum}: Success");
                                foreach (var item in parsed.Data)
                                {
                                    Console.WriteLine($"  - Title: {item.Title}");
                                }
                            }
                        }
                        catch (Exception ex)
                        {
                            Console.WriteLine($"Page {pageNum}: Success (JSON Parsing Failed: {ex.Message})");
                        }
                    }
                    else
                    {
                        Console.WriteLine($"Page {pageNum}: Failed ({result.Error?.Message})");
                    }
                }
            }
        }

First time
<img width="882" height="466" alt="dotnet1" src="https://github.com/bhilanisdk/media/blob/main/dotnet1.png" />
Second time
<img width="868" height="462" alt="dotnet2" src="https://github.com/bhilanisdk/media/blob/main/dotnet2.png" />
Third time
<img width="850" height="465" alt="dotnet3" src="https://github.com/bhilanisdk/media/blob/main/dotnet3.png" />

**🙏 Mata Shabari 🙏**
