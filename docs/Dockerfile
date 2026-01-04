FROM jekyll/jekyll:pages

WORKDIR /srv/jekyll

# Jekyll 3.x in this image needs webrick for `serve`.
RUN gem install webrick

EXPOSE 4000

# Fixed to GitHub Pages baseurl (/tmgc) for predictable local preview.
CMD ["sh", "-c", "jekyll serve --livereload --baseurl /tmgc"]
