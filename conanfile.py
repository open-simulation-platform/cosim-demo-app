from os import path
from conan import ConanFile
from conan.tools.files import copy


class ConanRecipe(ConanFile):
    name = "cosim-demo-app"
    url = "https://gitlab.sintef.no/open-simulation-platform/cosim-demo-app"
    settings = "os", "compiler", "build_type", "arch"
    package_type = "application"
    generators = "VirtualRunEnv"

    def configure(self):
        self.options["libcosim/*"].proxyfmu = True

    def requirements(self):
        self.requires("libcosimc/0.11.0@osp/testing-feature_conan-2")

    def generate(self):
        for dep in self.dependencies.values():
            if dep.ref.name == "libcosimc":
                copy(self, "cosim.h", dep.cpp_info.includedirs[0],
                     path.join(self.build_folder, "include"), keep_path=False)
            if dep.ref.name == "proxyfmu":
                copy(self, "proxyfmu*", dep.cpp_info.bindirs[0],
                     path.join(self.build_folder, "dist", "bin"), keep_path=False)
            for bindeps in dep.cpp_info.bindirs:
                copy(self, "*.dll", bindeps,
                     path.join(self.build_folder, "dist", "bin"), keep_path=False)

            for libdeps in dep.cpp_info.libdirs:
                copy(self, "*.so*", libdeps,
                     path.join(self.build_folder, "dist", "lib"), keep_path=False)