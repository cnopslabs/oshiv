# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Oshiv < Formula
  desc "Tool for finding and connecting to OCI instances"
  homepage "https://github.com/cnopslabs/oshiv"
  version "1.4.0"
  license "MIT"

  on_macos do
    url "https://github.com/cnopslabs/oshiv/releases/download/v1.4.0/oshiv_1.4.0_darwin_all.tar.gz"
    sha256 "ca8fe96cf728e620621f99ab4215b5cc8843ad33a3a1fdf8740919dd2e9b3962"

    def install
      bin.install "oshiv"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/cnopslabs/oshiv/releases/download/v1.4.0/oshiv_1.4.0_linux_amd64.tar.gz"
        sha256 "f06d6e3c7e2f5ff00c21c51421bb676ffd3e3f0b6420785ccca5e126db907ff0"

        def install
          bin.install "oshiv"
        end
      end
    end
    if Hardware::CPU.arm?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/cnopslabs/oshiv/releases/download/v1.4.0/oshiv_1.4.0_linux_arm64.tar.gz"
        sha256 "5bbdbbf005173a798e5bb1779a910e171889b56f908e446a08db49800e96c40d"

        def install
          bin.install "oshiv"
        end
      end
    end
  end
end
