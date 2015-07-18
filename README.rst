gosum
=====

Playing with go-routines to calculate different cksums at once.

I basicly try to do the "Parallel digestion" from
the Go Concurrency Patterns' `Pipelines and cancellation`_
on the Go blog.

.. _Pipelines and cancellation:
    https://blog.golang.org/pipelines

architecture
------------

  - for every file in argv:

    - ``main()`` opens input file
    - ``main()`` creates ``output_ch``
    - ``main()`` spawns ``read_fan(filehandle, alg_list, output_ch)``
      routine
    - ``main()`` starts listening for results on ``output_ch``
    - ``read_fan()`` calls ``hash_chan(alg, output_ch)`` for every
        algorithm: in ``alg_list``:

      - every ``hash_chan()`` spawns a ``chan_to_hash(input_chan,
        alg, output_ch)`` routine
      - every ``hash_chan()`` returns an ``input_chan`` for its
        ``chan_to_hash()``

    - ``read_fan()`` opens the file and starts replicating the data
      into all ``input_chan``'s
    - ``main()`` prints results from ``output_ch``
    - at EOF ``read_fan()`` closes all ``input_chan``'s and then
      ``output_ch``

  - after every file in argv is processed ``main()`` exits.

architecture 2.0
----------------
Idea is to use a single slice of bytes the producer reads
the input file to and signals the number of new bytes
to the consumers via a channel.

Roughly like this:

  - for every file in ``flag.Args()``:

    - ``main()`` opens input file
    - ``main()`` creates ``output_ch``
    - ``main()`` spawns ``read_fan(filehandle, alg_list, output_ch)``
      routine
    - ``main()`` starts listening for results on ``output_ch``
    - ``read_fan()`` calls ``hash_chan(alg, &buffer, output_ch)`` for
      every algorithm in ``alg_list``:

      - every ``hash_chan()`` spawns a ``chan_to_hash(input_chan,
        buffer_ptr, alg, output_ch)`` routine
      - every ``hash_chan()`` returns an ``input_chan`` for its
        ``chan_to_hash()``

    - ``read_fan()`` opens the file and starts writing the data
      into ``buffer``, notifying each ``chan_to_hash()`` via their
      ``input_chan``
    - ``main()`` prints results from ``output_ch``
    - at EOF ``read_fan()`` closes all ``input_chan``'s and then
      ``output_ch``

  - after every file in argv is processed ``main()`` exits.
