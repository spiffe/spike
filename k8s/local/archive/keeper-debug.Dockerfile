# Start with your built keeper image
FROM localhost:32000/spike-keeper:0.4.0 AS original

# Use Ubuntu as a base for better debugging capabilities
FROM ubuntu:22.04

# Copy the keeper binary from the original image
COPY --from=original /keeper /keeper

# Install basic debugging tools
RUN apt-get update && \
    apt-get install -y procps strace lsof curl file && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Set proper permissions and verify the binary
RUN chmod +x /keeper && \
    ls -la /keeper && \
    file /keeper

# Set the entrypoint to the keeper binary
ENTRYPOINT ["/keeper"]