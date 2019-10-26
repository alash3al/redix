class Redix < Formula
  desc "a persistent real-time key-value store, with the same redis protocol with powerful features"
  homepage "https://alash3al.github.io/redix/"
  url "https://github.com/alash3al/redix/releases/download/v1.10/redix_darwin_amd64.zip"
  sha256 "32511da3c6642aa2e22c178a50d90b6396f24ebeeb6867be45816d7328ad3e3d"

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
