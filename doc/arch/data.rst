.. Copyright (c) 2017 RackN Inc.
.. Licensed under the Apache License, Version 2.0 (the "License");
.. DigitalRebar Provision documentation under Digital Rebar master license
.. index::
  pair: DigitalRebar Provision; Data Architecture

.. _rs_data_architecture:

Data Architecture
=================

DigitalRebar Provision uses a fairly simple data model.  There are 4 main models for the provisioner server
and 4 models for the DHCP server.  Each model has a cooresponding API in the :ref:`rs_api`.

The models define elements of the system.  The API provides basic CRUD (create, read, update, and delete) operations as well as
some additional actions for manipulating the state of the system.  The :ref:`rs_api` contains that definitions of the actual
structures and methods on those objects.  Additionally, the :ref:`rs_operation` will describe common actions to use and do with
these models and how to build them.  The :ref:`rs_cli` describes the Command Line manipulators for the model.

This section will describe its use and role in the system.


.. _rs_provisioner_models:

Provisioner Models
------------------

These models represent things that the provisioner server use and manipulate.


.. index::
  pair: Model; Machine

.. _rs_model_machine:

Machine
~~~~~~~

The Machine object defines a machine that is being provisioned.  The Machine is represented by a unique **UUID**.  The UUID
is immutable after machine creation.  The machine's primary purpose is to map an incoming IP address to a :ref:`rs_model_bootenv`.
The :ref:`rs_model_bootenv` provides a set of rendered :ref:`rs_model_template` that will can be used to boot the machine.  The
machine provides parameters to the :ref:`rs_model_template`.  The Machine provides configuration to the renderer in the
form of parameters and fields.  The **Name** field should contain the FQDN of the node.

The Machine object contains an **Error** field that represents errors encountered while operating on the machine.  In general,
these are errors pertaining to rendering the :ref:`rs_model_bootenv`.

The Machine parameters are defined as a field on the Machine that is presented as a dicitionary of string keys to arbritary objects.
These could be strings, bools, numbers, arrays, or objects represented similarly defined dictionaries.  The machine parameters
are available to templates for expansion in them.

.. index::
  pair: Model; BootEnv

.. _rs_model_bootenv:

BootEnv
~~~~~~~

The BootEnv object defines an environment to boot a machine.  It has two main components an OS information section and a templates
list.  The OS information section defines what makes up the installation base for this bootenv.  It defines the install ISO, a
URL to get the ISO, and SHA256 checksum to validate the image.  These are used to provide the basic install image, kernel, and
base packages for the bootenv.

The other primary section is a set of templates that present files in the file server's name space that can served via HTTP or
TFTP.  The templates can be in-line in the BootEnv object or reference a :ref:`rs_model_template`.  The templates are specified as
a list of paths in the filesystem and either an ID of a :ref:`rs_model_template` or inline content.  The path field of the
template information can use the same template expansion that is used in the template.  See :ref:`rs_model_template` for more
information.

Additionally, the BootEnv defines required and optional parameters.  The required parameters validated at render time to be
present or an error is generated.  These parameters can be met by the parameters on the machine or from the global :ref:`rs_model_param`
space.

BootEnvs can be marked **OnlyUnknown**.  This tells the rest of the system that this BootEnv is not for specific machines.  It is a
general BootEnv.  For example, *discovery* and *ignore* are **OnlyUnknown**.  *discovery* is used to discovery unknown machines and
add them to DigitalRebar Provision.  *ignore* is a special bootenv that tells machines to boot their local disk.  These BootEnvs
populate the pxelinux.0, ipxe, and elilo default fallthrough files.  These are different than their counterpart BootEnvs,
*sledgehammer* and *local* which are machine specific BootEnvs that populate configuration files that are specific to a single
machine.  A machine boots *local*; an unknown machine boots *ignore*.  There can only be one **OnlyUnknown** BootEnv active
at a time.  This is specified by the :ref:`rs_model_prefs` *unknownBootEnv*.


.. index::
  pair: Model; Template

.. _rs_model_template:

Template
~~~~~~~~

The Template object defines a templated content that can be referenced by its ID.  The content of the template (or
in-line template in a :ref:`rs_model_bootenv`) is a `golang text/template <https://golang.org/pkg/text/template/#hdr-Actions>`_ string.
The template has a set of special expansions.  The normal expansion syntax is:

  ::

    {{ .Machine.Name }}

This would expand to the machine's **Name** field.  There are helpers for the parameter spaces, the :ref:`rs_model_bootenv` object,
and some miscellaneous functions.  Additionally, the normal `golang text/template <https://golang.org/pkg/text/template/#hdr-Actions>`_
functions are available as well.  Things like **range**, **len**, and comparators are available as well.  Currently, **template** inclusion
is not supported.

The following table lists the current set of expansion custom functions:

============================== =================================================================================================================================================================================================
Expansion                      Description
============================== =================================================================================================================================================================================================
.Machine.Name                  The FQDN of the Machine in the Machine object stored in the **Name** field
.Machine.ShortName             The Name part of the FDQN of the Machine object stored in the **Name** field
.Machine.UUID                  The Machine's **UUID** field
.Machine.Path                  A path to a custom machine unique space in the file server name space.
.Machine.Address               The **Address** field of the Machine
.Machine.HexAddress            The **Address** field of the Machine in Hex format (useful for elilo config files
.Machine.URL                   A HTTP URL that references the Machine's specific unique filesystem space.
.Env.PathFor <proto> <file>    This references the boot environment and builds a string that presents a either a tftp or http specifier into exploded ISO space for that file.  *Proto* is **tftp** or **http**.  The *file* is a relative path inside the ISO.
.Env.InstallURL                An HTTP URL to the base ISO install directory.
.Env.JoinInitrds <proto>       A comma separated string of all the initrd files specified in the BootEnv reference through the specified proto (**tftp** or **http**)
.BootParams                    This renders the **BootParam** field of :ref:`rs_model_bootenv` at that spot.  Template expansion applies to that field as well.
.ProvisionerAddress            An IP address that is on the provisioner that is the most direct access to the machine.
.ProvisionerURL                An HTTP URL to access the base file server root
.ApiURL                        An HTTPS URL to access the DigitalRebar Provision API
.GenerateToken                 This generates limited use access token for the machine to either update itself if it exists or create a new machine.  The token's validity is limited in time by global preferences.  See :ref:`rs_model_prefs`.
.ParseURL <segment> <url>      Parse the specified URL and return the segment requested.
.ParamExists <key>             Returns true if the specified key is a valid parameter available for this rendering.
.Param <key>                   Returns the structure for the specified key for this rendering.
============================== =================================================================================================================================================================================================

**GenerateToken** is very special.  This generates either a *known token* or an *unknown token* for use by the template to update objects
in DigitalRebar Provision.  The tokens are valid for a limited time as defined by the **knownTokenTimeout** and **unknownTokenTimeout**
preferences respectively.  The tokens are also restricted to the function the can perform.  The *known token* is limited to only
reading and updating the specific machine the template is being rendered for.  If a machine is not present during the render, an
*unknown token* is generated that has the ability to query and create machines.  These are used by the install process to indicate that
the install is finished and that the *local* BootEnv should be used for the next boot and during the discovery process to create
the newly discovered machine.

.. note::
  **.Machine.Path** is particularly useful for ensure that templates are expanded into a unique file space for
  each machine.  An example of this is per machine kickstart files.  These can be seen in the **assets/bootenvs/ubuntu-16.04.yml**.

With regard to the **.Param** and **.ParamExists** functions, these functions return the parameter or existence of
the parameter specified by the *key* input.  The parameters are examined from most specific to global.  This means
that the Machine object's parameters are checked first, then the global :ref:`rs_model_param`.  The parameters on machines
and the global space are free form dictionaries and default empty.  Any key/value pair can be added and referenced.

The default :ref:`rs_model_template` and :ref:`rs_model_bootenv` use the following optional (unless marked with an \*)
parameters.

=================================  ================  =================================================================================================================================
Parameter                          Type              Description
=================================  ================  =================================================================================================================================
ntp_servers                        Array of objects  lookup format
proxy-servers                      Array of objects  lookup format
operating-system-disk              String            A string to use as the default install drive.  /dev/sda or sda depending upon kickstart or preseed.
access_keys                        Map of strings    The key is the name of the public key.  The value is the public key.  All keys are placed in the .authorized_keys file of root.
provisioner-default-password-hash  String            The password hash for the initial default password, **RocketSkates**
provisioner-default-user           String            The initial user to create for ubuntu/debian installs
dns-domain                         String            DNS Domain to use for this system's install
\*operating-system-license-key     String            Windows Only
\*operating-system-install-flavor  String            Windows Only
=================================  ================  =================================================================================================================================

For some examples of this in use, see :ref:`rs_operation`.

.. index::
  pair: Model; Param

.. _rs_model_param:

Param
~~~~~

The Param Object defines a single key / value.  The system maintains a global list of these manipulated by the :ref:`rs_api`.  The key space
is a free form string and the value is an arbirtary data blob specified by JSON through the :ref:`rs_api`.  The common parameters defined
in :ref:`rs_model_template` can be globally set through these objects.  They are the lowest level of precedence.

.. _rs_dhcp_models:

DHCP Models
-----------

These models represent things that the DHCP server use and manipulate.

.. index::
  pair: Model; Subnet

.. _rs_model_subnet:

Subnet
~~~~~~

The Subnet Object defines the configuration of a single subnet for the DHCP server to process.  Multiple subnets are allowed.  The Subnet
can represent a local subnet attached to an interface local (Broadcast Subnet) to the DigitalRebar Provision server or a subnet that is
being forwarded or relayed (Relayed Subnet) to the DigitalRebar Provision server.

The subnet is uniquely identified by its **Name**.  The subnet defines a CIDR-based range with a specific subrange to hand out for
nodes that do NOT have explicit reservations (**ActiveStart** thru **ActiveEnd**).  The subnet also defines the **NextServer** in
the PXE chain.  This is usually an IP associated with DigitalRebar Provision, but if the provisioner is disabled, this can be
any next hop server.  The leases for both reserved and unreserved clients as specified here (**ReservedLeaseTime** and **ActiveLeaseTime**).
The subnet can also me marked as only working for explicitly reserved nodes (**ReservedOnly**).

The subnet also allows for the specification of DHCP options to be sent to clients.  These can be overriden by :ref:`rs_model_reservation`
specific options.  Some common options are:

========  ====  =================================
Type      #     Description
========  ====  =================================
IP        3     Default Gateway
IP        6     DNS Server
IP        15    Domain Name
String    67    Next Boot File - e.g. lpxelinux.0
========  ====  =================================

golang template expansion also works in these fields.  This can be used to make custom request based reply options.

For example, this value in the Next Boot File option (67) will return a file based upon what type of machine is booting.  If
the machine supports, iPXE then an iPXE boot image is sent, if the system is marked for legacy bios, then lpxelinux.0 is returned,
otherwise return a 64-bit UEFI network boot loader:

  ::

    {{if (eq (index . 77) "iPXE") }}default.ipxe{{else if (eq (index . 93) "0")}}lpxelinux.0{{else}}bootx64.efi{{end}}


The data element for the template expansion as represented by the '.' above is a map of strings indexed by an integer.  The
integer is the option number from the DHCP request incoming options.  The IP addresses and other data fields are converted to
a string form (dotted quads or base 10 numerals).

The final elements of a subnet are the **Strategy** and **Pickers** options.  These are described in the :ref:`rs_api` JSON description.
They define how a node should be identified (**Strategy**) and the algorithm for picking addresses (**Pickers**)).  The strategy can
only be set to **MAC** currently.  This will use the MAC address of the node as its DHCP identifier.  Others may show up in time.

**Pickers** defines an order lists of method to determine the address to hand out.  Currently, this will default to the list:
*hint*, *nextFree*, and *mostExpired*.  The following options are available for the list.

* hint - Use what was provided in the DHCP Offer/Request
* nextFree - Within the subnet's pool of Active IPs, choose the next free making sure to loop over all addresses before reuse.
* mostExpired - If no free address is available, use the most expired address first.
* none - Do NOT hand out anything


.. index::
  pair: Model; Reservation

.. _rs_model_reservation:

Reservation
~~~~~~~~~~~

The Reservation Object defines a mapping between a token and an IP address.  The token is defined by the assigned strategy.  Similar
to :ref:`rs_model_subnet`, the only current strategy is **MAC**.  This will use the MAC address of the incoming requests as the
identity token.  The reservation allows for the optional specification of specific options and a next server that override or
augment the options defined in a subnet.  Because the reservation is an explicit binding of the token to an IP address, the
address can be handed out with the definition of a subnet.  This requires that the reservation have the Netmask Option (Option 1)
specified.  In general, it is a good idea to define a subnet that will cover the reservation with default options and parameters, but
it is not required.

.. index::
  pair: Model; Lease

.. _rs_model_lease:

Lease
~~~~~

The Lease Object defines the ephemeral mapping of a token, as defined by the reservation or subnets strategy, and an IP address assigned
by the reservation or pulled form the subnet's pool.  The lease contains the Strategy used for the token and the experation time.  The
contents of the lease are immutable with the exception of the expiration time.

.. index::
  pair: Model; Interface

.. _rs_model_interface:

Interface
~~~~~~~~~

The Interface Object is a read-only object that is used to identify local interfaces and their addresses on the
DigitalRebar Provision server.  This is useful for determing what subnets to create and with what address ranges.
The :ref:`rs_ui_subnets` part of the :ref:`rs_ui` uses this to populate possible subnets to create.


.. _rs_additional_models:

Additional Models
-----------------

These models control additional parts and actions of the system.

.. index::
  pair: Model; User

.. _rs_model_user:

User
~~~~

The User Object controls access to the system.  The User object contains a name and a password hash for validating access.  Additionally,
the User :ref:`rs_api` can be used to generate time-based, function restricted tokens for use in :ref:`rs_api` calls.  The
:ref:`rs_model_template` provides a helper function to generate these for restricted machine access in the discovery and post-install
process.

More on access tokens and an control in :ref:`rs_operation`.

.. index::
  pair: Model; Prefs

.. _rs_model_prefs:

Prefs
~~~~~

Most configuration is handle through the global parameter system on machines specific parameters, but there are a few modifiable
options that can be changed over time in the server (outside of command line flags).  These are preferences.  The preferences are
key value pairs where both the key and the value are string.  The use internally may be an integer, but the specification through
the :ref:`rs_api` is by string.

=================== ======= ==================================================================================================================================================================================
Pref                Type    Description
=================== ======= ==================================================================================================================================================================================
defaultBootEnv      string  This is a valid :ref:`rs_model_bootenv` the is assign to a :ref:`rs_model_machine` if the machine does not have a bootenv specified.  The default is **sledgehammer**.
unknownBootEnv      string  This is the :ref:`rs_model_bootenv` used when a boot request is serviced by an unknown machine.  The BootEnv must have **OnlyUnknown** set to true.  The default is **ignore**.
unknownTokenTimeout integer The amount of time in seconds that the token generated by **GenerateToken** is valid for unknown machines.  The default is 600 seconds.
knownTokenTimeout   integer The amount of time in seconds that the token generated by **GenerateToken** is valid for known machines.  The default is 3600 seconds.
=================== ======= ==================================================================================================================================================================================

.. _rs_special_objects:

Special Objects
---------------

These are not objects in the system but represent files and directories in the server space.

.. index::
  pair: Model; Files

.. _rs_model_file:

Files
~~~~~

File server has a managed filesystem space.  The :ref:`rs_api` defines methods to upload, destroy, and get these files outside of the
normal TFTP and HTTP path.  The TFTP and HTTP access paths are read-only.  The only way to modify this space is through the :ref:`rs_api`
or direct filesystem access underneath DigitalRebar Provision.  The filesystem space defaults to */var/lib/tftpboot*, but can be overridden
by the command line flag *--file-root*, e.g. --file-root=`pwd`/drp-data when using --isolated on install.  These directories can be
directly manipulated by administrators for faster loading times.

This space is also used by the :ref:`rs_model_bootenv` import process when "exploding" an ISO for use by :ref:`rs_model_machine`.

.. index::
  pair: Model; Isos

.. _rs_model_iso:

Isos
~~~~

The ISO directory in the file server space is managed specially by the ISO :ref:`rs_api`.  The API handles upload and destroy
functionality.  The API also handles notification of the BootEnv system to "explode" ISOs that are needed by BootEnvs and marking
the BootEnv as available.

ISOs can be directly placed into the **isos** directory in the file root, but the using BootEnvs need to be modified or deleted and
re-added to force the ISO to be exploded for use.
