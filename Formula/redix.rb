class Redix < Formula
  desc "a persistent real-time key-value store, with the same redis protocol with powerful features"
  homepage "https://alash3al.github.io/redix/"
  url "https://github.com/alash3al/redix/releases/download/v1.6/redix_darwin_amd64.zip"
  sha256 "2ee32559f97f57e3274d7fcbdf2a580925b5c6e826d241f545c1799ba88d31fc"

  def install
    # Rename redix_darwin_amd64 To Redix
    system "mv", "redix_darwin_amd64", "redix"
    # Install the package very important step
    bin.install "redix"
  end

  test do
    system "true"
  end
end
