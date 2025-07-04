# Sample Dockerfile for testing fws
FROM nginx:alpine

# Copy a simple index.html
COPY index.html /usr/share/nginx/html/

# Add a health check endpoint
RUN echo '#!/bin/sh\necho "Container is healthy"' > /healthcheck.sh && \
    chmod +x /healthcheck.sh

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /healthcheck.sh

EXPOSE 80

# Add custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

CMD ["nginx", "-g", "daemon off;"] 