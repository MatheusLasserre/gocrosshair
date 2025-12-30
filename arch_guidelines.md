When building packages for Arch Linux, adhere to the package guidelines below, especially if the intention is to contribute a new package to Arch Linux. You should also see the PKGBUILD(5) and makepkg(8) man pages.

Important points listed on this page are not repeated on the other package guideline pages. These specific guidelines are intended as an addition to the standards listed below.

Packages submitted to the Arch User Repository must additionally comply with AUR submission guidelines.

See .proto files in the /usr/share/pacman/ directory as PKGBUILD examples.

Package etiquette
Packages should never be installed to /usr/local/.
Do not introduce new variables or functions into PKGBUILD build scripts, unless the package cannot be built without doing so, as these could possibly conflict with variables and functions used in makepkg itself.
If a new variable or a new function is absolutely required, prefix its name with an underscore (_), e.g.
_customvariable=
Avoid using /usr/libexec/ for anything. Use /usr/lib/$pkgname/ instead.
The packager field from the package meta file can be customized by the package builder by modifying the appropriate option in the /etc/makepkg.conf file, or alternatively override it by creating ~/.makepkg.conf.
Do not use makepkg subroutines (e.g. error, msg, msg2, plain, warning) as they might change at any time. To print data, use printf or echo.
All important messages should be echoed during install using an .install file. For example, if a package needs extra setup to work, directions should be included.
Dependencies are the most common packaging error. Please take the time to verify them carefully, for example by running ldd on dynamic executables, checking tools required by scripts or looking at the documentation of the software. The namcap utility can help you in this regard. This tool can analyze both PKGBUILD and the resulting package tarball and will warn you about bad permissions, missing dependencies, redundant dependencies, and other common mistakes.
Any optional dependencies that are not needed to run the package or have it generally function should not be included in the depends array; instead the information should be added to the optdepends array:
optdepends=('cups: printing support'
            'sane: scanners support'
            'libgphoto2: digital cameras support'
            'alsa-lib: sound support'
            'giflib: GIF images support'
            'libjpeg: JPEG images support'
            'libpng: PNG images support')
The above example is taken from the wine package. The optdepends information is automatically printed out on installation/upgrade so one should not keep this kind of information in .install files.
When creating a package description for a package, do not include the package name in a self-referencing way. For example, "Nedit is a text editor for X11" could be simplified to "A text editor for X11". Also try to keep the descriptions to ~80 characters or less.
Try to keep the line length in the PKGBUILD below ~100 characters.
Where possible, remove empty lines from the PKGBUILD (provides, replaces, etc.)
It is common practice to preserve the order of the PKGBUILD fields in the same order as given in the PKGBUILD article. However, this is not mandatory, as the only requirement in this context is correct Bash syntax.
Quote variables which may contain spaces, such as "$pkgdir" and "$srcdir".
To ensure the integrity of packages, make sure that the integrity variables contain correct values. These can be updated using the updpkgsums(8) tool.
Package naming
Package names can contain only alphanumeric characters and any of @, ., _, +, -. Names are not allowed to start with hyphens or dots. All letters should be lowercase.
Package names should not be suffixed with the upstream major release version number (e.g. we do not want libfoo2 if upstream calls it libfoo v2.3.4) in case the library and its dependencies are expected to be able to keep using the most recent library version with each respective upstream release. However, for some software or dependencies, this can not be assumed. In the past this has been especially true for widget toolkits such as GTK and Qt. Software that depends on such toolkits can usually not be trivially ported to a new major version. As such, in cases where software can not trivially keep rolling alongside its dependencies, package names should carry the major version suffix (e.g. gtk2, gtk3, qt4, qt5). For cases where most dependencies can keep rolling along the newest release but some cannot (for instance closed source that needs libpng12 or similar), a deprecated version of that package might be called libfoo1 while the current version is just libfoo.
Package versioning
Package version — pkgver — should be the same as the version released by the author.
Versions can include letters if need be, e.g. version could be 2.54BETA32.
Version tags may not include hyphens, and may contain letters, numbers, and periods only. If the upstream version contains a hyphen, it must be replaced with an underscore.

Package releases — pkgrel — are specific to Arch Linux packages. These allow users to differentiate between newer and older package builds. When a new package version is first released, the release count starts at 1. Then as fixes and optimizations are made, the package will be re-released to the Arch Linux public and the release number will increment.
When a new version comes out, the release count resets to 1.
Package release tags follow the same naming restrictions as version tags.
Package dependencies
Do not rely on transitive dependencies in any of the PKGBUILD#Dependencies, as they might break, if one of the dependencies is updated.
List all direct library dependencies. To identify them find-libdeps(1) (part of devtools) can be used.
Package relations
Do not add $pkgname to PKGBUILD#provides, as it is always implicitly provided by the package.
Do not add $pkgname to PKGBUILD#conflicts, as a package cannot conflict with itself.
List all external shared libraries of a package in PKGBUILD#provides (e.g. 'libsomething.so'). To identify them find-libprovides(1) (part of devtools) can be used.
Package sources
HTTPS sources (https:// for tarballs, git+https:// for git sources) should be used wherever possible
Sources should be verified using PGP signatures wherever possible (this might entail building from a git tag instead of a source tarball, if upstream signs commits and tags but not the tarballs)
The factual accuracy of this article or section is disputed.

Reason: commit# is not required in recent pacman as proper checksum is supported for git sources. See[1]. gitea package also has been updated to use below approach (Discuss in Talk:Arch package guidelines)
When building from a git tag, use its object hash obtained from git rev-parse instead of the tag name:
_tag=1234567890123456789012345678901234567890 # git rev-parse "v$pkgver"
source=(git+https://$url.git?signed#tag=$_tag)

pkgver() {
    cd "$pkgname"
    git describe
}
An example for this approach can be found in the gitea package. The reason for this practice is that tags can be force pushed to change the commit that they are pointing to, which would alter the built package. Using the tag object hash ensures the integrity of the sources because force pushing the tag changes its hash. Using a pkgver() function prevents accidentally bumping pkgver without updating _tag as well. See VCS package guidelines#VCS sources for more info on the formatting of VCS sources.
Do not diminish the security or validity of a package (e.g. by removing a checksum check or by removing PGP signature verification), because an upstream release is broken or suddenly lacks a certain feature (e.g. PGP signature missing for a new release)
Sources have to be unique in srcdir (this might require renaming them when downloading, e.g. "${pkgname}-${pkgver}.tar.gz::https://${pkgname}.tld/download/${pkgver}.tar.gz")
Avoid using specific mirrors (e.g. on sourceforge) to download, as they might become unavailable
Git objects (e.g., tags, commits, etc.) signed by an SSH key can be verified using a git command with gpg.ssh.allowedSignersFile pointing to a file specifying possible signing keys. See [2] for an example.
Working with upstream
It is considered best-practice to work closely with upstream wherever possible. This entails reporting problems about building and testing a package.

Report problems to upstream right away.
Upstream patches wherever possible.
Add comments with links to relevant (upstream) bug tracker tickets in the PKGBUILD (this is particularly important, as it ensures, that other packagers can understand changes and work with a package as well).
It is recommended to track upstream with tools such as nvchecker, nvrsAUR or urlwatch to be informed about new stable releases.

Directories
Configuration files should be placed in the /etc directory. If there is more than one configuration file, it is customary to use a subdirectory in order to keep the /etc area as clean as possible. Use /etc/pkg where pkg is the name of the package (or a suitable alternative, eg, apache uses /etc/httpd/).
Package files should follow these general directory guidelines:
/etc	System-essential configuration files
/usr/bin	Binaries
/usr/lib	Libraries
/usr/include	Header files
/usr/lib/pkg	Modules, plugins, etc.
/usr/share/doc/pkg	Application documentation
/usr/share/info	GNU Info system files
/usr/share/licenses/pkg	Application licenses
/usr/share/man	Manpages
/usr/share/pkg	Application data
/var/lib/pkg	Persistent application storage
/etc/pkg	Configuration files for pkg
/opt/pkg	Large self-contained packages
Packages should not contain any of the following directories:
/bin
/sbin
/dev
/home
/srv
/media
/mnt
/proc
/root
/selinux
/sys
/tmp
/var/tmp
/run
Makepkg duties
When makepkg is used to build a package, it does the following automatically:

Checks if package dependencies and makedepends are installed
Downloads source files from servers
Checks the integrity of source files
Unpacks source files
Does any necessary patching
Builds the software and installs it in a fake root
Strips symbols from binaries
Strips debugging symbols from libraries
Compresses manual and/or info pages
Generates the package meta file which is included with each package
Compresses the fake root into the package file
Stores the package file in the configured destination directory (i.e. the current working directory by default)
Architectures
The arch array should contain 'x86_64' if the compiled package is architecture-specific. Otherwise, use 'any' for architecture independent packages.

Licenses
There are two kinds of licenses regarding an Arch package:

PKGBUILD's license field
The license field of a PKGBUILD. It lists the packaged software's upstream license. It is NOT the license of the package source. The licenses in this field must be in the SPDX license format. See also PKGBUILD#license for more details.

Package sources licenses
The license for the package sources themselves. In RFC40, Arch Linux specifies that package sources are to be licensed as 0BSD with RFC52 specifying that REUSE should be used to enforce this.

It boils down to this:

Have a LICENSE file in the sources root with exactly this content. This is Arch Linux's 0BSD license for packages.
Have a REUSE.toml in the sources root. You can use pkgctl license setup to generate a reasonable config to get you started.
Make sure to run pkgctl license check and that it returns no errors.
If you have additional files that you need to license, you need to pick a reasonable license for them. This is usually quite straight forward:

If the file in question (for example, a launcher script launcher.sh or a systemd service file myunit.service) was created entirely by you or other Arch staff, license it as 0BSD.
Note
If there is a patch that you wrote that you also want to submit upstream, you can still license it as 0BSD for Arch and allow upstream to apply their license on submission.
If the file was taken from upstream (for instance, an icon tool.png or a patch fix.patch), then it should carry the upstream license.
See also the Arch Linux Dev blog post introducing pkgctl-license(1).

Reproducible builds
Arch is working on making all packages reproducible. A packager can check if a package is reproducible with makerepropkg from devtools or repro from archlinux-repro.

$ makerepropkg $pkgname-1-1-any.pkg.tar.zst
Or

$ repro -f $pkgname-1-1-any.pkg.tar.zst
If the timestamp is required at build-time, use the environment variable SOURCE_DATE_EPOCH. The format is documented upstream.

--- PKG BUILD GUIDE LINES

This article discusses variables definable by the maintainer in a PKGBUILD. For information on the PKGBUILD functions and creating packages in general, refer to Creating packages. Also read PKGBUILD(5).

A PKGBUILD is a Bash script containing the build information required by Arch Linux packages.

Packages in Arch Linux are built using the makepkg utility. When makepkg is run, it searches for a PKGBUILD file in the current directory and follows the instructions therein to either compile or otherwise acquire the files to build a package archive—pkgname.pkg.tar.zst. The resulting package contains binary files and installation instructions, readily installable with pacman.

Mandatory variables are pkgname, pkgver, pkgrel, and arch. license is not strictly necessary to build a package, but is recommended for any PKGBUILD shared with others, as makepkg will produce a warning if not present.

It is a common practice to define the variables in the PKGBUILD in the same order as given here. However, it is not mandatory.

Tip
Use namcap to check PKGBUILDs for common packaging mistakes.
Use shellcheck(1) to check PKGBUILDs for common scripting mistakes. See also SC2034, SC2154 and SC2164:
shellcheck --shell=bash --exclude=SC2034,SC2154,SC2164 PKGBUILD
termux-language-serverAUR provides a language server for PKGBUILD, makepkg.conf, etc.
See the .proto files in the /usr/share/pacman/ directory as examples.

Package name
pkgbase
When building regular packages, this variable should not be explicitly declared in the PKGBUILD: its value defaults to that of #pkgname.

When building a split package, this variable can be used to explicitly specify the name to be used to refer to the group of packages in the output of makepkg and in the naming of source-only tarballs. The value is not allowed to begin with a hyphen. If not specified, the value will default to the first element in the pkgname array.

All options and directives for split packages default to the global values given in the PKGBUILD. Nevertheless, the following ones can be overridden within each split package’s packaging function: #pkgdesc, #arch, #url, #license, #groups, #depends, #optdepends, #provides, #conflicts, #replaces, #backup, #options, #install, and #changelog.

pkgname
Either the name of the package, e.g. pkgname=foo, or, for split packages, an array of names, e.g. pkgname=(foo bar). Package names should only consist of lowercase alphanumerics and the following characters: @._+- (at symbol, dot, underscore, plus, hyphen). Names are not allowed to start with hyphens or dots. For the sake of consistency, pkgname should match the name of the source tarball of the software: for instance, if the software is in foobar-2.5.tar.gz, use pkgname=foobar.

Version
pkgver
The version of the package. This should be the same as the version published by the author of the upstream software. It can contain letters, numbers, periods and underscores, but not a hyphen (-). If the author of the software uses one, replace it with an underscore (_). If the pkgver variable is used later in the PKGBUILD, then the underscore can easily be substituted for a hyphen, e.g. source=("${pkgname}-${pkgver//_/-}.tar.gz").

Note
If upstream uses a timestamp versioning such as 30102014, ensure to use the reversed date, i.e. 20141030 (ISO 8601 format). Otherwise it will not appear as a newer version.
Tip
The ordering of uncommon values can be tested with vercmp(8), which is provided by the pacman package.
makepkg can automatically update this variable by defining a pkgver() function in the PKGBUILD. See VCS package guidelines#The pkgver() function for details.
pkgrel
The release number. This is usually a positive integer number that allows to differentiate between consecutive builds of the same version of a package. As fixes and additional features are added to the PKGBUILD that influence the resulting package, the pkgrel should be incremented by 1. When a new version of the software is released, this value must be reset to 1. In exceptional cases other formats can be found in use, such as major.minor.

epoch
Warning
epoch should only be used when absolutely required to do so.
Used to force the package to be seen as newer than any previous version with a lower epoch. This value is required to be a non-negative integer; the default is 0. It is used when the version numbering scheme of a package changes (or is alphanumeric), breaking normal version comparison logic. For example:

pkgver=5.13
pkgrel=2
epoch=1
1:5.13-2
See pacman(8) for more information on version comparisons.

Generic
pkgdesc
The description of the package. This is recommended to be 80 characters or less and should not include the package name in a self-referencing way, unless the application name differs from the package name. For example, use pkgdesc='Text editor for X11' instead of pkgdesc='Nedit is a text editor for X11'.

Also it is important to use keywords wisely to increase the chances of appearing in relevant search queries.

arch
An array of architectures that the PKGBUILD is intended to build and work on. Arch officially supports only x86_64, but other projects may support other architectures. For example, Arch Linux 32 provides support for i686 and pentium4, and Arch Linux ARM provides support for armv7h (armv7 hardfloat) and aarch64 (armv8 64-bit).

There are two types of values the array can use:

arch=(any) indicates the package can be built on any architecture, and once built, is architecture-independent in its compiled state (usually shell scripts, fonts, themes, many types of extensions, Java programs, etc.).
arch=(...) with one or more architectures (but not any) indicates the package can be compiled for any of the specified architectures, but is architecture-specific once compiled. For these packages, specify all architectures that the PKGBUILD officially supports. For official repository and AUR packages, this means arch=('x86_64'). Optionally, AUR packages may choose to additionally support other known working architectures.
The target architecture can be accessed with the variable CARCH during a build.

url
The URL of the official site of the software being packaged.

license
This article or section is a candidate for merging with Arch package guidelines#licenses.

Notes: The PKGBUILD format does not enforce a packaging policy. (Discuss in Talk:PKGBUILD)
This article or section needs expansion.

Reason: Add more details from [1]. (Discuss in Talk:PKGBUILD)
The license under which the software is distributed. Arch Linux uses SPDX license identifiers. Each license must have a corresponding entry in /usr/share/licenses/.

For common licenses (like GPL-3.0-or-later), package licenses delivers all the corresponding files. The package is installed by default, as it is a dependency of base meta package, and the files may be found in /usr/share/licenses/spdx/. Simply refer to the license using its SPDX license identifier from the list of SPDX identifiers.

License families like BSD or MIT are, strictly speaking, not a single license and each instance requires a separate license file. In license variable refer to them using a common SPDX identifier (e.g. BSD-3-Clause or MIT), but then provide the corresponding file as if it was a custom license.

For custom licenses the identifier should be either LicenseRef-license-name or custom:license-name, if they are not covered by the common families mentioned above. The corresponding license text must be placed in directory /usr/share/licenses/pkgname. To install the file a following code snippet may be used in package() section:

install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
Note
pkgdir variable is defined by makepkg, see PKGBUILD(5) § PACKAGING FUNCTIONS for more information.
Combining multiple licenses or adding exceptions should follow the SPDX syntax. For example a package released under either GNU/GPL 2.0 or GNU/LGPL 2.1 could use 'GPL-2.0-or-later OR LGPL-2.1-or-later', a package released under Apache 2.0 with LLVM exception would use 'Apache-2.0 WITH LLVM-exception' and a package released with part under the BSD 3 clause, others under GNU/LGPL 2.1 and some under GNU/GPL 2.0 would use 'BSD-3-Clause AND LGPL-2.1-or-later AND GPL-2.0-or-later'[2]. Note that this must be a single string, so the entire expression has to be enclosed in quotes. As for November 2023 SPDX list of exceptions is limited, so usually the custom license route must be used.

If issues are encountered with SPDX identifiers, during the transitional period using old identifiers —names of the directories in /usr/share/licenses/common— is acceptable.

See also Nonfree applications package guidelines.

Additional information and perspectives on free and open source software licenses may be found on the following pages:

Wikipedia:Free software license
Wikipedia:Comparison of free and open-source software licenses
A Legal Issues Primer for Open Source and Free Software Projects
GNU Project - Various Licenses and Comments about Them
Debian - License information
Open Source Initiative - Licenses by Name
groups
The group the package belongs in. For instance, when installing plasma, it installs all packages belonging in that group.

Dependencies
Note
Additional architecture-specific arrays can be added by appending an underscore and the architecture name, e.g. optdepends_x86_64=().
depends
An array of packages that must be installed for the software to build and run. Dependencies defined inside the package() function are only required to run the software.

Version restrictions can be specified with comparison operators, e.g. depends=('foobar>=1.8.0'); if multiple restrictions are needed, the dependency can be repeated for each, e.g. depends=('foobar>=1.8.0' 'foobar<2.0.0').

This article or section is a candidate for merging with Arch package guidelines.

Notes: The PKGBUILD format does not enforce a packaging policy. (Discuss in Talk:PKGBUILD)
The depends array should list all direct first level dependencies even when some are already declared transitively. For instance, if a package foo depends on both bar and baz, and the bar package depends in turn on baz too, it will ultimately lead to undesired behavior if bar stops pulling in baz. Pacman will not enforce the installation of baz on systems which newly install the foo package, or have cleaned up orphans, and foo will crash at runtime or otherwise misbehave.

In some cases this is not necessary and may or may not be listed, for example glibc cannot be uninstalled as every system needs some C library, or python for a package that already depends on another python- module, as the second module must per definition depend on python and cannot ever stop pulling it in as a dependency.

Dependencies should normally include the requirements for building all optional features of a package. Alternatively, any feature whose dependencies are not included should be explicitly disabled via a configure option. Failure to do this can lead to packages with "automagic dependencies" build-time optional features that are unpredictably enabled due to transitive dependencies or unrelated software installed on the build machine, but which are not reflected in the package dependencies.

If the dependency name appears to be a library, e.g. depends=(libfoobar.so), makepkg will try to find a binary that depends on the library in the built package and append the soname version needed by the binary. Manually appending the version disables automatic detection, e.g. depends=('libfoobar.so=2').

makedepends
An array of packages that are only required to build the package. The minimum dependency version can be specified in the same format as in the depends array. The packages in the depends array are implicitly required to build the package, they should not be duplicated here.

Note
The package base-devel is assumed to be already installed when building with makepkg. Dependencies of this package should not be included in makedepends array.
If using VCS sources, do not forget to include the appropriate VCS tool (git, subversion, cvs, ...).
Tip
Use pactree -rsud1 package | grep base-devel to check whether a particular package is a direct dependency of base-devel (requires pacman-contrib).
checkdepends
An array of packages that the software depends on to run its test suite, but are not needed at runtime. Packages in this list follow the same format as depends. These dependencies are only considered when the check() function is present and is to be run by makepkg.

Note
The package base-devel is assumed to be already installed when building with makepkg. Dependencies of this package should not be included in checkdepends array.
optdepends
An array of packages that are not needed for the software to function, but provide additional features. This may imply that not all executables provided by a package will function without the respective optdepends.[3] If the software works on multiple alternative dependencies, all of them can be listed here, instead of the depends array.

A short description of the extra functionality each optdepend provides should also be noted:

optdepends=('cups: printing support'
            'sane: scanners support'
            'libgphoto2: digital cameras support'
            'alsa-lib: sound support'
            'giflib: GIF images support'
            'libjpeg: JPEG images support'
            'libpng: PNG images support')
Package relations
Note
Additional architecture-specific arrays can be added by appending an underscore and the architecture name, e.g. conflicts_x86_64=().
provides
An array of additional packages that the software provides the features of, including virtual packages such as cron or sh and all external shared libraries. Packages providing the same item can be installed side-by-side, unless at least one of them uses a conflicts array.

Note
The version that the package provides should be mentioned (pkgver and potentially the pkgrel), in case packages referencing the software require one. For instance, a modified qt package version 3.3.8, named qt-foobar, should use provides=('qt=3.3.8'); omitting the version number would cause the dependencies that require a specific version of qt to fail.
Do not add pkgname to the provides array, as it is done automatically.
conflicts
An array of packages that conflict with, or cause problems with the package, if installed. All these packages and packages providing this item will need to be removed. The version properties of the conflicting packages can also be specified in the same format as the depends array.

Note that conflicts are checked against pkgname as well as names specified in the provides array. Hence, if your package provides a foo feature, specifying foo in the conflicts array will cause a conflict between your package and all other packages that contain foo in their provides array (i.e., there is no need to specify all those conflicting package names in your conflicts array). Let us take a concrete example:

netbeans implicitly provides netbeans as the pkgname itself
A hypothetical netbeans-cpp package would provide netbeans and conflicts with netbeans
A hypothetical netbeans-php package would provide netbeans and conflicts with netbeans, but does not need to explicitly conflict with netbeans-cpp since packages providing the same feature are implicitly in conflict.
When packages provide the same feature via the provides array, there is a difference between explicitly adding the alternative package to the conflicts array and not adding it. If the conflicts array is explicitly declared the two packages providing the same feature will be considered as alternative; if the conflicts array is missing the two packages providing the same feature will be considered as possibly cohabiting. Packagers should always ignore the content of the provides variable in deciding whether to declare a conflicts variable or not.

replaces
An array of obsolete packages that are replaced by the package, e.g. wireshark-qt uses replaces=('wireshark'). When syncing, pacman will immediately replace an installed package upon encountering another package with the matching replaces in the repositories. If providing an alternate version of an already existing package or uploading to the AUR, use the conflicts and provides arrays, which are only evaluated when actually installing the conflicting package.

Others
backup
An array of files that can contain user-made changes and should be preserved during upgrade or removal of a package, primarily intended for configuration files in /etc. If these files are unchanged from how they ship with the package, they will be removed or replaced as normal files during upgrade or removal.

Files in this array should use relative paths without the leading slash (/) (e.g. etc/pacman.conf, instead of /etc/pacman.conf). The backup array does not support empty directories or wildcards such as "*".

When updating, new versions may be saved as file.pacnew to avoid overwriting a file which already exists and was previously modified by the user. Similarly, when the package is removed, user-modified files will be preserved as file.pacsave unless the package was removed with the pacman -Rn command.

See also Pacnew and Pacsave files.

options
This array allows overriding some of the default behavior of makepkg, defined in /etc/makepkg.conf. To set an option, include the name in the array. To disable an option, place an ! before it.

The full list of the available options can be found in PKGBUILD(5) § OPTIONS AND DIRECTIVES.

install
The name of the .install script to be included in the package.

pacman has the ability to store and execute a package-specific script when it installs, removes or upgrades a package. The script contains the following functions which run at different times:

pre_install — The script is run right before files are extracted. One argument is passed: new package version.
post_install — The script is run right after files are extracted. Any additional notes that should be printed after the package is installed should be located here. One argument is passed: new package version.
pre_upgrade — The script is run right before files are extracted. Two arguments are passed in the following order: new package version, old package version.
post_upgrade — The script is run right after files are extracted. Two arguments are passed in the following order: new package version, old package version.
pre_remove — The script is run right before files are removed. One argument is passed: old package version.
post_remove — The script is run right after files are removed. One argument is passed: old package version.
Each function is run chrooted inside the pacman install directory. See this thread.

Tip
A prototype .install is provided at /usr/share/pacman/proto.install.
pacman#Hooks provide similar functionality.
Note
Do not end the script with exit. This would prevent the contained functions from executing.
changelog
The name of the package changelog. To view changelogs for installed packages (that have this file):

$ pacman -Qc pkgname
Sources
source
An array of files needed to build the package. It must contain the location of the software source, which in most cases is a full HTTP or FTP URL. The previously set variables pkgname and pkgver can be used effectively here; e.g. source=("https://example.com/${pkgname}-${pkgver}.tar.gz").

Files can also be supplied in the same directory where the PKGBUILD is located, and their names added to this array. Before the actual build process starts, all the files referenced in this array will be downloaded or checked for existence, and makepkg will not proceed if any is missing.

.install files are recognized automatically by makepkg and should not be included in the source array. Files in the source array with extensions .sig, .sign, or .asc are recognized by makepkg as PGP signatures and will be automatically used to verify the integrity of the corresponding source file.

Warning
The downloaded source filename must be unique because the SRCDEST directory can be the same for all packages. For instance, using the version number of the project as a filename potentially conflicts with other projects with the same version number. In this case, the alternative unique filename to be used is provided with the syntax source=('unique_package_name::file_uri'); e.g. source=("${pkgname}-${pkgver}.tar.gz::https://github.com/coder/program/archive/v${pkgver}.tar.gz").
Tip
Additional architecture-specific arrays can be added by appending an underscore and the architecture name, e.g. source_x86_64=(). There must be a corresponding integrity array with checksums, e.g. sha256sums_x86_64=().
Some servers restrict download by filtering the User-Agent string of the client or other types of restrictions, which can be circumvented with DLAGENTS.
Use file:// URL to point to a directory or a file in your computer filesystem. For example, a local Git repository can be specified as "${pkgname}::git+file:///path/to/repository".
Magnet link support can be added using transmission-dlagentAUR as DLAGENT and using the magnet:// URI prefix instead of the canonical magnet:?.
See PKGBUILD(5) § USING VCS SOURCES and VCS package guidelines#VCS sources for details on VCS specific options, such as targeting a specific Git branch or commit.
noextract
An array of files listed under source, which should not be extracted from their archive format by makepkg. This can be used with archives that cannot be handled by /usr/bin/bsdtar or those that need to be installed as-is. If an alternative unarchiving tool is used (e.g. lrzip), it should be added in the makedepends array and the first line of the prepare() function should extract the source archive manually; for example:

prepare() {
  lrzip -d source.tar.lrz
}
Note that while the source array accepts URLs, noextract is just the file name portion:

source=("http://foo.org/bar/foobar.tar.xz")
noextract=('foobar.tar.xz')
To extract nothing, consider the following:

If source contains only plain URLs without custom file names, strip the source array before the last slash:
noextract=("${source[@]##*/}")
If source contains only entries with custom file names, strip the source array after the :: separator (taken from a previous version of firefox-i18n's PKGBUILD):
noextract=("${source[@]%%::*}")
Note
If an archive has many top-level files, there is a risk of unwanted overwriting files extracted from other source archives. If that is the case, consider adding it to noextract and extracting it manually into a subdirectory.
validpgpkeys
An array of PGP fingerprints. If used, makepkg will only accept signatures from the keys listed here and will ignore the trust values from the keyring. If the source file was signed with a subkey, makepkg will still use the primary key for comparison.

Only full fingerprints are accepted. They must be uppercase and must not contain whitespace characters.

Note
Use gpg --list-keys --fingerprint KEYID to find out the fingerprint of the appropriate key.
Please read makepkg#Signature checking for more information.

Integrity
These variables are arrays whose items are checksum strings that will be used to verify the integrity of the respective files in the source array. Insert SKIP for a particular file, and its checksum will not be tested.

The checksum type and values should always be those provided by upstream, such as in release announcements. When multiple types are available, the strongest checksum is to be preferred (in order from most to least preferred): b2, sha512, sha384, sha256, sha224, sha1, md5, ck. This best ensures the integrity of the downloaded files, from upstream announcement to package building.

Note
Additionally, when upstream makes digital signatures available, the signature files should be added to the source array and the PGP key fingerprint to the validpgpkeys array. This allows authentication of the files at build time.
The values for these variables can be auto-generated by makepkg's -g/--geninteg option, then commonly appended with makepkg -g >> PKGBUILD. The updpkgsums(8) command from pacman-contrib is able to update the variables wherever they are in the PKGBUILD. Both tools will use the variable that is already set in the PKGBUILD, or fall back to sha256sums if none is set.

The file integrity checks to use can be set up with the INTEGRITY_CHECK option in /etc/makepkg.conf. See makepkg.conf(5).

Note
Additional architecture-specific arrays can be added by appending an underscore and the architecture name, e.g. sha256sums_x86_64=().
b2sums
An array of BLAKE2b checksums with digest size of 512 bits.

sha512sums, sha384sums, sha256sums, sha224sums
An array of SHA-2 checksums with digest sizes 512, 384, 256 and 224 bits, respectively. sha256sums is the most common of them.

sha1sums
An array of 160-bit SHA-1 checksums of the files listed in the source array.

md5sums
An array of 128-bit MD5 checksums of the files listed in the source array.

cksums
An array CRC32 checksums (from UNIX-standard cksum) of the files listed in the source array.


--- SUBMISSION GUIDELINES

Users can share PKGBUILD scripts using the Arch User Repository. It does not contain any binary packages but allows users to upload PKGBUILDs that can be downloaded by others. These PKGBUILDs are completely unofficial and have not been thoroughly vetted, so they should be used at your own risk.

Submitting packages
Warning
Before attempting to submit a package you are expected to familiarize yourself with Arch package guidelines and PKGBUILD.
Verify carefully that what you are uploading is correct. Packages that violate the rules may be deleted without warning.
If you are unsure in any way about the package or the build/submission process even after reading this section twice, submit the PKGBUILD to the AUR mailing list, the AUR forum on the Arch forums, or ask on our IRC channel for public review before adding it to the AUR.

Rules of submission
When submitting a package to the AUR, observe the following rules:

The submitted PKGBUILDs must not build applications already in any of the official binary repositories under any circumstances. Check the official package database for the package. If any version of it exists, do not submit the package. If the official package is out-of-date, flag it as such. If the official package is broken or is lacking a feature, then please file a bug report.
Exception to this strict rule may only be packages having extra features enabled and/or patches in comparison to the official ones. In such an occasion the pkgname should be different to express that difference. For example, a package for GNU screen containing the sidebar patch could be named screen-sidebar. Additionally the conflicts=('screen') array should be used in order to avoid conflicts with the official package.
Check the AUR if the package already exists. If it is currently maintained, changes can be submitted in a comment for the maintainer's attention. If it is unmaintained or the maintainer is unresponsive, the package can be adopted and updated as required. Do not create duplicate packages.
Make sure the package you want to upload is useful. Will anyone else want to use this package? Is it extremely specialized? If more than a few people would find this package useful, it is appropriate for submission.
The AUR and official repositories are intended for packages which install general software and software-related content, including one or more of the following: executable(s); configuration file(s); online or offline documentation for specific software or the Arch Linux distribution as a whole; media intended to be used directly by software.
Packages that do not support the x86_64 architecture are not allowed in the AUR.
Do not use replaces in an AUR PKGBUILD unless the package is to be renamed, for example when Ethereal became Wireshark. If the package is an alternate version of an already existing package, use conflicts (and provides if that package is required by others). The main difference is: after syncing (-Sy) pacman immediately wants to replace an installed, 'offending' package upon encountering a package with the matching replaces anywhere in its repositories; conflicts, on the other hand, is only evaluated when actually installing the package, which is usually the desired behavior because it is less invasive.
Packages that build from a version control system and are not tied to a specific version need to have an appropriate suffix, -git for git and so on, see VCS package guidelines#Package naming for a full list.
Packages that use prebuilt deliverables, when the sources are available, must use the -bin suffix. An exception to this is with Java. The AUR should not contain the binary tarball created by makepkg, nor should it contain the filelist. If you are packaging a non-free software, see also Nonfree applications package guidelines#Package naming regarding usage of -bin suffix.
Packages that build from source using a specific version do not use a suffix.
Please add a comment line to the top of the PKGBUILD file which contains information about the current maintainers and previous contributors, respecting the following format. Remember to disguise your email to protect against spam. Additional lines are optional.
Note
The use of obfuscation when it comes to email addresses makes it difficult for people to contact you.
If you are assuming the role of maintainer for an existing PKGBUILD, add your name to the top like this
# Maintainer: Your Name <address at domain dot tld>
If there were previous maintainers, put them as contributors. The same applies for the original submitter if this is not you. If you are a co-maintainer, add the names of the other current maintainers as well.
# Maintainer: Your name <address at domain dot tld>
# Maintainer: Other maintainer's name <address at domain dot tld>
# Contributor: Previous maintainer's name <address at domain dot tld>
# Contributor: Original submitter's name <address at domain dot tld>
Add a LICENSE file and/or a REUSE.toml file to your repository. You are encouraged to follow the Arch package guidelines#Package sources licenses and license your submission under the 0BSD license.
Note
Packages missing a license or containing a different license than 0BSD are not eligible for promotion to the official repositories.
Authentication
For write access to the AUR, you need to have an SSH key pair. The content of the public key needs to be copied to your profile in My Account, and the corresponding private key configured for the aur.archlinux.org host. For example:

~/.ssh/config
Host aur.archlinux.org
  IdentityFile ~/.ssh/aur
  User aur
You should create a new key pair rather than use an existing one, so that you can selectively revoke the keys should something happen:

$ ssh-keygen -f ~/.ssh/aur
Tip
You can add multiple public keys to your profile by separating them with a newline in the input field.
Creating package repositories
If you are creating a new package from scratch, establish a local Git repository and an AUR remote by cloning the intended pkgbase. If the package does not yet exist, the following warning is expected:

$ git -c init.defaultBranch=master clone ssh://aur@aur.archlinux.org/pkgbase.git
Cloning into 'pkgbase'...
warning: You appear to have cloned an empty repository.
Checking connectivity... done.
Note
The repository will not be empty if pkgbase matches a deleted package.
If you already have a package, initialize it as a Git repository if it is not one:

$ git -c init.defaultBranch=master init
and add an AUR remote:

$ git remote add label ssh://aur@aur.archlinux.org/pkgbase.git
Then fetch this remote to initialize it in the AUR.

Note
Pull and rebase to resolve conflicts if pkgbase matches a deleted package.
Publishing new package content
Warning
Your commits will be authored with your global Git name and email address. It is very difficult to change commits after pushing them (FS#45425). If you want to push to the AUR under different credentials, you can change them per package with git config user.name "..." and git config user.email "...".
When releasing a new version of the packaged software, update the pkgver or pkgrel variables to notify all users that an upgrade is needed. Do not update those values if only minor changes to the PKGBUILD such as the correction of a typo are being published.

Do not commit mere pkgver bumps for VCS packages. They are not considered out of date when the upstream has new commits. Only do a new commit when other changes are introduced, such as changing the build process.

Be sure to regenerate .SRCINFO whenever PKGBUILD metadata changes, such as pkgver() updates; otherwise the AUR will not show updated version numbers.

To upload or update a package:

add at least PKGBUILD and .SRCINFO,
add any additional new or modified helper files (such as .install files or local source files such as patches),
add a package source license,
commit with a meaningful commit message,
push the changes to the AUR.
For example:

$ makepkg --printsrcinfo > .SRCINFO
$ git add PKGBUILD .SRCINFO
$ git commit -m "useful commit message"
$ git push
Note
If .SRCINFO was not included in your last commit, add it by changing your last commit with git commit --amend so the AUR will permit your push.
The AUR only allows pushes to the master branch. If the local branch is named something else, rename it and push again.
Tip
To keep the working directory and commits as clean as possible, create a gitignore(5) that excludes all files and force-add files as needed.
Maintaining packages
Check for feedback and comments from other users and try to incorporate any improvements they suggest; consider it a learning process!
Please do not leave a comment containing the version number every time you update the package. This keeps the comment section usable for valuable content mentioned above.
Please do not just submit and forget about packages! It is the maintainer's job to maintain the package by checking for updates and improving the PKGBUILD.
If you do not want to continue to maintain the package for some reason, disown the package using the AUR web interface and/or post a message to the AUR Mailing List. If all maintainers of an AUR package disown it, it will become an "orphaned" package.
Automation is a valuable tool for maintainers, but it can not replace manual intervention (e.g. projects can change license, add or remove dependencies, and other notable changes even for "minor" releases). Automated PKGBUILD updates are used at your own risk and any malfunctioning accounts and their packages may be removed without prior notice.
Requests
Deletion, merge, and orphan requests can be created by clicking on the "Submit Request" link under "Package Actions" on the right hand side. This dispatches notification emails to the current package maintainer and to the aur-requests mailing list for discussion. Package Maintainers will then either accept or reject the request.

Deletion
Request to unlist a pkgbase from the AUR. A short note explaining the reason for deletion is required, as well as supporting details (like when a package is provided by another package, if you are the maintainer yourself, it is renamed and the original owner agreed, etc).

Note
It is not sufficient to explain why a package is up for deletion only in its comments: as soon as a package maintainer takes action, the only place where such information can be obtained is the aur-requests mailing list.
Deletion requests can be rejected, in which case if you are the maintainer you will likely be advised to disown the package to allow adoption by another maintainer.
After a package is "deleted", its git repository remains available for cloning.
Merge
Request to delete a pkgbase and transfer its votes and comments to another pkgbase. The name of the package to merge into is required.

This is the action to use if, for example, an upstream has renamed their project.

Note
This has nothing to do with git merge or GitLab's merge requests.
As the transfer of the votes and comments requires an already existing destination, if a package has no votes or comments, a deletion request linking to the new package is identical.
Orphan
Request that a pkgbase be disowned. These requests will be granted after two weeks if the current maintainer did not react. The exception is if a package was flagged out-of-date for at least 180 days; orphan requests are then automatically accepted.